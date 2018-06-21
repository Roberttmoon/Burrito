package main

import (
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
        '$ burito env_var env_var1 env_var2

Each env_var should be the SSM Parameter-Store key that you are setting. 
Burrito will output burrito.sh which contains a shell script that will
set the value of each variable to the returned SSM Parameter-Store value.

`)
}

func get_parameter(parameter_name string) string {
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
	}
	value := *param.Parameter.Value
	return value
}

func main() {
	ssm_parameter_variables := os.Args[1:]
	number_of_params := len(ssm_parameter_variables)

	if number_of_params == 0 {
		help()
		fmt.Printf("Error: No parameters were passed in, please include some parameters\n\n")
		os.Exit(1)
	}

	file_string := "#!/bin/bash\n\n"

	for env_var := 0; env_var < number_of_params; env_var++ {
		parameter_key := os.Getenv(ssm_parameter_variables[env_var])
		parameter_value := get_parameter(parameter_key)
		file_string = fmt.Sprintf("%vexport %v=%v\n", file_string, ssm_parameter_variables[env_var], parameter_value)
	}

	file_out := []byte(file_string)

	err := ioutil.WriteFile("burito.sh", file_out, 0644)

	if err != nil {
		fmt.Printf("Error: Could not write file burito.sh: %v", err)
		os.Exit(2)
	}
}
