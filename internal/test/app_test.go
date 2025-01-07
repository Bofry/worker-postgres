package test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Bofry/config"
	postgres "github.com/Bofry/worker-postgres"
	"github.com/joho/godotenv"
)

var (
	__TEST_JAEGER_TRACE_URL string

	__ENV_FILE        = ".env"
	__ENV_FILE_SAMPLE = ".env.sample"

	__CONFIG_YAML_FILE        = "config.yaml"
	__CONFIG_YAML_FILE_SAMPLE = "config.yaml.sample"
)

type MessageManager struct {
	GoTestTopic *GoTestSlotMessageHandler `slot:"gotestSlot"`
	Invalid     *InvalidMessageHandler
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func TestMain(m *testing.M) {
	var err error

	_, err = os.Stat(__CONFIG_YAML_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			err = copyFile(__CONFIG_YAML_FILE_SAMPLE, __CONFIG_YAML_FILE)
			if err != nil {
				panic(err)
			}
		}
	}

	_, err = os.Stat(__ENV_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			err = copyFile(__ENV_FILE_SAMPLE, __ENV_FILE)
			if err != nil {
				panic(err)
			}
		}
	}

	{
		f, err := os.Open(__ENV_FILE)
		if err != nil {
			panic(err)
		}
		env, err := godotenv.Parse(f)
		if err != nil {
			panic(err)
		}
		__TEST_JAEGER_TRACE_URL = env["TEST_JAEGER_TRACE_URL"]
	}

	godotenv.Load()
	m.Run()
}

func TestStartup(t *testing.T) {
	app := App{}
	starter := postgres.Startup(&app).
		Middlewares(
			postgres.UseMessageManager(&MessageManager{}),
			postgres.UseErrorHandler(func(ctx *postgres.Context, msg *postgres.Message, err interface{}) {
				t.Logf("catch err: %v", err)
			}),
			postgres.UseTracing(false),
		).
		ConfigureConfiguration(func(service *config.ConfigurationService) {
			service.
				LoadEnvironmentVariables("").
				LoadYamlFile("config.yaml").
				LoadCommandArguments()

			t.Logf("%+v\n", app.Config)
		})

	runCtx, cancel := context.WithTimeout(context.Background(), 13*time.Second)
	defer cancel()
	if err := starter.Start(runCtx); err != nil {
		t.Error(err)
	}

	select {
	case <-runCtx.Done():
		if err := starter.Stop(context.Background()); err != nil {
			t.Error(err)
		}
	}

	// assert app.Config
	{
		// conf := app.Config
		// var expectedNsqAddress string = os.Getenv("TEST_NSQLOOKUPD_ADDRESS")
		// if conf.NsqAddress != expectedNsqAddress {
		// 	t.Errorf("assert 'Config.NsqAddress':: expected '%v', got '%v'", expectedNsqAddress, conf.NsqAddress)
		// }
	}
}
