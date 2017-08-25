package main

import (
	"fmt"
	"os"
	"sort"

	ini "github.com/copland/iniparse/pkg/iniparse"
	"github.com/mitchellh/go-homedir"
	cli "gopkg.in/urfave/cli.v1"
)

type awsProfiles []*ini.Section

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

func loadAwsCredentials(credsPath string) (awsProfiles, error) {

	if credsPath == "" {
		dir, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		credsPath = fmt.Sprintf("%s/.aws/credentials", dir)
	}
	awsCreds, err := ini.NewIniFile(credsPath)
	if err != nil {
		return nil, fmt.Errorf("error: could not load %s", credsPath)
	}
	awsProfiles := awsProfiles(awsCreds.Sections)
	return awsProfiles, nil

}

func (profiles awsProfiles) activateProfile(profileToActivate string) error {
	for _, profile := range profiles {
		if profile.Name == profileToActivate {
			if !profile.KeyIsPresent("aws_access_key_id") {
				return fmt.Errorf("error: profile is missing aws_access_key_id")
			}
			if !profile.KeyIsPresent("aws_secret_access_key") {
				return fmt.Errorf("error: profile is missing aws_secret_access_key")
			}
			fmt.Printf("export AWS_DEFAULT_PROFILE=%s\n", profile.Name)
			fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", profile.Keys["aws_access_key_id"])
			fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", profile.Keys["aws_secret_access_key"])
			return nil
		}
	}
	return fmt.Errorf("error: could not find profile %s\n", profileToActivate)
}

var list = cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "list profiles in AWS credentials file",
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
	Name:    "activate",
	Aliases: []string{"a"},
	Usage:   "activate AWS profile",
	Action: func(c *cli.Context) error {
		awsProfiles, err := loadAwsCredentials("")
		if err != nil {
			return err
		}
		maybeErr := awsProfiles.activateProfile(c.Args().First())
		return maybeErr
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
	}

	app.Run(os.Args)

}
