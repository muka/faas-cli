package api

import (
	"fmt"
	"sync"

	"github.com/openfaas/faas-cli/builder"
	"github.com/openfaas/faas-cli/options"
	"github.com/openfaas/faas-cli/stack"
)

//Push a function to repository
func Push(arg options.PushOptions) error {

	var services stack.Services
	if arg.Services != nil {
		services = *arg.Services
	} else {
		if len(arg.YamlFile) > 0 {
			parsedServices, err := stack.ParseYAMLFile(arg.YamlFile, arg.Regex, arg.Filter)
			if err != nil {
				return err
			}

			if parsedServices != nil {
				services = *parsedServices
			}
		}
	}

	if len(services.Functions) > 0 {
		pushStack(&services, arg.Parallel)
	} else {
		return fmt.Errorf("you must supply a valid YAML file")
	}
	return nil
}

func pushImage(image string) {
	builder.ExecCommand("./", []string{"docker", "push", image})
}

func pushStack(services *stack.Services, queueDepth int) {
	wg := sync.WaitGroup{}

	workChannel := make(chan stack.Function)

	for i := 0; i < queueDepth; i++ {

		go func(index int) {
			wg.Add(1)
			for function := range workChannel {
				fmt.Printf("[%d] > Pushing: %s.\n", index, function.Name)
				if len(function.Image) == 0 {
					fmt.Println("Please provide a valid Image value in the YAML file.")
				} else {
					pushImage(function.Image)
				}
			}

			fmt.Printf("[%d] < Pushing done.\n", index)
			wg.Done()
		}(i)
	}

	for k, function := range services.Functions {
		function.Name = k
		workChannel <- function
	}

	close(workChannel)

	wg.Wait()

}
