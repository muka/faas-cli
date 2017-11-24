package builder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/openfaas/faas-cli/options"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/faas-cli/templates"
)

//Build one or more functions
func Build(arg options.BuildOptions) error {

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

	if pullErr := PullTemplates(""); pullErr != nil {
		return fmt.Errorf("could not pull templates for OpenFaaS: %v", pullErr)
	}

	if len(services.Functions) > 0 {
		return BuildStack(&services, arg.Parallel, arg.Nocache, arg.Squash, arg.Shrinkwrap)
	}

	if len(arg.Image) == 0 {
		return fmt.Errorf("please provide a valid --image name for your Docker image")
	}
	if len(arg.Handler) == 0 {
		return fmt.Errorf("please provide the full path to your function's handler")
	}
	if len(arg.FunctionName) == 0 {
		return fmt.Errorf("please provide the deployed --name of your function")
	}
	if err := BuildImage(
		arg.Image,
		arg.Handler,
		arg.FunctionName,
		arg.Language,
		arg.Nocache,
		arg.Squash,
		arg.Shrinkwrap,
	); err != nil {
		return err
	}

	return nil
}

//BuildStack build a stack of functions
func BuildStack(services *stack.Services, queueDepth int, nocache bool, squash bool, shrinkwrap bool) error {
	wg := sync.WaitGroup{}

	workChannel := make(chan stack.Function)
	var buildErr error
	for i := 0; i < queueDepth; i++ {

		go func(index int) {
			// interrupt build
			if buildErr != nil {
				return
			}
			wg.Add(1)
			for function := range workChannel {
				fmt.Printf("[%d] > Building: %s.\n", index, function.Name)
				if len(function.Language) == 0 {
					fmt.Println("Please provide a valid --lang or 'Dockerfile' for your function.")

				} else {
					if err := BuildImage(
						function.Image,
						function.Handler,
						function.Name,
						function.Language,
						nocache,
						squash,
						shrinkwrap,
					); err != nil {
						buildErr = err
					}
				}
			}

			fmt.Printf("[%d] < Builder done.\n", index)
			wg.Done()
		}(i)
	}

	for k, function := range services.Functions {
		if buildErr != nil {
			break
		}
		if function.SkipBuild {
			fmt.Printf("Skipping build of: %s.\n", function.Name)
		} else {
			function.Name = k
			workChannel <- function
		}
	}

	close(workChannel)

	wg.Wait()

	return buildErr
}

// PullTemplates pulls templates from Github from the master zip download file.
func PullTemplates(templateURL string) error {
	var err error
	exists, err := os.Stat(filepath.Join(os.Getenv("workdir"), "./template"))
	if err != nil || exists == nil {
		log.Println("No templates found in current directory.")

		err = templates.PullTemplates(templateURL)
		if err != nil {
			log.Println("Unable to download templates from Github.")
			return err
		}
	}
	return err
}
