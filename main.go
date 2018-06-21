package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func help() {
	fmt.Print(`
Burrito collects ssm parameters and wraps them up... like a burrito!

Collect your parameters by calling:
        '$ Burrito env_var env_var1 env_var2

Each env_var should be the SSM Parameter-Store key that you are setting. 
Burrito will output burrito.sh which contains a shell script that will
set the value of each variable to the returned SSM Parameter-Store value.

`)
}

func get_parameter(parameter_name string) (string, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("us-east-1")},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		fmt.Print("Failed to make a new session: %v\n", err)
		os.Exit(3)
	}

	ssmsvc := ssm.New(sess, aws.NewConfig().WithRegion("us-east-1"))
	key_name := parameter_name
	withDecryption := true

	param, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &key_name,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		fmt.Printf("Error: Could not find parameter: %v\n", err)
		return "", err
	}
	value := *param.Parameter.Value
	return value, nil
}

type cli_options struct {
	file_path   string
	file_header string
}

func collect_cli_options() *cli_options {
	cli_opt := cli_options{}
	flag.StringVar(&cli_opt.file_path, "file_path", "burrito.sh", "This is the filepath for the output script, default is 'burrito.sh'")
	flag.StringVar(&cli_opt.file_header, "file_header", "#!/bin/bash", "This is the fisrt line of the output script, defualt is '#!/bin/bash'")

	return &cli_opt
}

func parameter_args() []string {
	env_vars := flag.Args()

	if len(env_vars) == 0 {
		help()
		fmt.Printf("Error: No parameters were passed in, please include some parameters\n\n")
		os.Exit(1)
	}

	return env_vars
}

func main() {
	cli_opts := collect_cli_options()
	flag.Parse()
	ssm_parameter_variables := parameter_args()
	number_of_params := len(ssm_parameter_variables)

	file_string := cli_opts.file_header

	for env_var := 0; env_var < number_of_params; env_var++ {
		parameter_key := os.Getenv(ssm_parameter_variables[env_var])
		parameter_value, err := get_parameter(parameter_key)

		if err != nil {
			fmt.Printf("Skipped parameter: %v\n", ssm_parameter_variables[env_var])
			continue
		}

		file_string = fmt.Sprintf("%vexport %v=%v\n", file_string, ssm_parameter_variables[env_var], parameter_value)
	}

	file_out := []byte(file_string)

	err := ioutil.WriteFile(cli_opts.file_path, file_out, 0644)

	if err != nil {
		fmt.Printf("Error: Could not write file burrito.sh: %v", err)
		os.Exit(2)
	}
}
