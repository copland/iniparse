package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/mitchellh/go-homedir"
	cli "gopkg.in/urfave/cli.v1"

	ini "github.com/copland/iniparse/pkg/iniparse"
)

type awsProfiles []*ini.Section
type awsProfile struct {
	data ini.Section
}

// Len returns the count of sections
func (profiles awsProfiles) Len() int {
	return len(profiles)
}

// Less determines whether one section is
// alphabetically less than another
func (profiles awsProfiles) Less(i, j int) bool {
	return profiles[i].Name < profiles[j].Name
}

// Swap changes the index of two sections
// in the profiles
func (profiles awsProfiles) Swap(i, j int) {
	profiles[i], profiles[j] = profiles[j], profiles[i]
}

func (profiles awsProfiles) getProfile(name string) (*awsProfile, error) {
	for _, profile := range profiles {
		if profile.Name == name {
			return &awsProfile{data: *profile}, nil
		}
	}
	return nil, fmt.Errorf("error: could not find profile %s", name)
}

func loadAwsCredentials(credsPath string) (awsProfiles, error) {

	if credsPath == "" {
		dir, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		credsPath = fmt.Sprintf("%s/.aws/credentials", dir)
	}
	awsCreds := ini.IniFile{}
	err := awsCreds.Load(credsPath)
	if err != nil {
		return nil, fmt.Errorf("error: could not load %s", credsPath)
	}
	awsProfiles := awsProfiles(awsCreds.Sections)
	return awsProfiles, nil
}

func dumpAwsCredentials(awsProfiles awsProfiles, credsPath string) error {
	if credsPath == "" {
		dir, err := homedir.Dir()
		if err != nil {
			return err
		}
		credsPath = fmt.Sprintf("%s/.aws/credentials", dir)
	}
	iniFile := ini.IniFile{Path: credsPath, Sections: awsProfiles}
	err := iniFile.Dump()
	if err != nil {
		return err
	}
	return nil
}

func (profile awsProfile) activate() error {
	if !profile.data.KeyIsPresent("aws_access_key_id") {
		return fmt.Errorf("error: profile is missing aws_access_key_id")
	}
	if !profile.data.KeyIsPresent("aws_secret_access_key") {
		return fmt.Errorf("error: profile is missing aws_secret_access_key")
	}
	fmt.Printf("export AWS_DEFAULT_PROFILE=%s\n", profile.data.Name)
	fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", profile.data.Keys["aws_access_key_id"])
	fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", profile.data.Keys["aws_secret_access_key"])
	return nil
}

func (profile awsProfile) update(user string) error {
	sess := session.Must(session.NewSession())
	iamSvc := iam.New(sess)
	input := iam.CreateAccessKeyInput{UserName: &user}
	output, err := iamSvc.CreateAccessKey(&input)
	if err != nil {
		return err
	}
	currentCreds, err := iamSvc.Config.Credentials.Get()
	deleteAccessKeyInput := iam.DeleteAccessKeyInput{UserName: &user, AccessKeyId: &currentCreds.AccessKeyID}
	iamSvc.DeleteAccessKey(&deleteAccessKeyInput)
	profile.data.Keys["aws_access_key_id"] = *output.AccessKey.AccessKeyId
	profile.data.Keys["aws_secret_access_key"] = *output.AccessKey.SecretAccessKey
	return nil
}

var list = cli.Command{
	Name:  "list",
	Usage: "list profiles in AWS credentials file",
	Action: func(c *cli.Context) error {
		awsProfiles, err := loadAwsCredentials("")
		if err != nil {
			return err
		}
		sort.Sort(awsProfiles)
		for _, profile := range awsProfiles {
			fmt.Printf("%s\n", profile.Name)
		}
		return nil
	},
}

var activate = cli.Command{
	Name:  "activate",
	Usage: "activate AWS profile",
	Action: func(c *cli.Context) error {
		awsProfiles, err := loadAwsCredentials("")
		if err != nil {
			return err
		}
		profileToActivate, err := awsProfiles.getProfile(c.Args().First())
		if err != nil {
			return err
		}
		maybeErr := profileToActivate.activate()
		return maybeErr
	},
}

var update = cli.Command{
	Name:  "update",
	Usage: "generate new AWS access/secret key pair for user in profile",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Usage: "AWS user to create new key pair for",
		},
		cli.StringFlag{
			Name:  "profiles, p",
			Usage: "comma-separated list of profiles to update",
		},
	},
	Action: func(c *cli.Context) error {
		awsProfiles, err := loadAwsCredentials("")
		if err != nil {
			return err
		}
		var user string
		if c.String("user") != "" {
			user = c.String("user")
		} else if c.String("u") != "" {
			user = c.String("user")
		} else {
			return fmt.Errorf("error: user not set")
		}
		var profiles string
		if c.String("profiles") != "" {
			profiles = c.String("profiles")
		} else if c.String("p") != "" {
			profiles = c.String("p")
		} else {
			return fmt.Errorf("error: profiles not set")
		}
		for _, profileName := range strings.Split(profiles, ",") {

			profile, err := awsProfiles.getProfile(profileName)
			fmt.Printf("Updating %s...", profile.data.Name)
			if profile.data.KeyIsPresent("aws_access_key_id") {
				os.Setenv("AWS_ACCESS_KEY_ID", profile.data.Keys["aws_access_key_id"])
			} else {
				return fmt.Errorf("error: aws_access_key_id is missing")
			}
			if profile.data.KeyIsPresent("aws_secret_access_key") {
				os.Setenv("AWS_SECRET_ACCESS_KEY", profile.data.Keys["aws_secret_access_key"])
			} else {
				return fmt.Errorf("error: aws_secret_access_key is missing")
			}
			if err != nil {
				return err
			}
			err = profile.update(user)
			if err != nil {
				fmt.Printf("SKIPPING\n")
				fmt.Printf("%s\n", err.Error())
			} else {
				os.Setenv("AWS_ACCESS_KEY_ID", "")
				os.Setenv("AWS_SECRET_ACCESS_KEY", "")
				fmt.Printf("DONE\n")
			}
		}
		err = dumpAwsCredentials(awsProfiles, "")
		if err != nil {
			return err
		}
		return nil
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "awscreds"
	app.Usage = "Easily manage AWS credentials files"
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		list,
		activate,
		update,
	}

	app.Run(os.Args)

}
