package skeleton

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"app/generated/engine"

	bima "github.com/bimalabs/framework/v4"
	"github.com/bimalabs/framework/v4/configs"
	"github.com/bimalabs/framework/v4/drivers"
	"github.com/bimalabs/framework/v4/events"
	"github.com/bimalabs/framework/v4/interfaces"
	"github.com/bimalabs/framework/v4/middlewares"
	"github.com/bimalabs/framework/v4/parsers"
	"github.com/bimalabs/framework/v4/routes"
	"github.com/fatih/color"
	"github.com/goccy/go-json"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/grpclog"
	"gopkg.in/yaml.v2"
)

type (
	Application string
)

func (_ Application) Run(config string) {
	if config == "" {
		config = ".env"
	}

	container, err := engine.NewContainer(bima.Application)
	if err != nil {
		panic(err)
	}

	env := container.GetBimaConfig()
	loadEnv(env, config, filepath.Ext(config))

	workDir, _ := os.Getwd()

	var ext []string
	var cName bytes.Buffer

	ext = parsers.ParseModule(workDir)
	servers := make([]configs.Server, 0, len(ext))
	for _, c := range ext {
		cName.Reset()
		cName.WriteString(c)
		cName.WriteString(":server")

		servers = append(servers, container.Get(cName.String()).(configs.Server))
	}

	ext = parsers.ParseListener(workDir)
	listeners := make([]events.Listener, 0, len(ext))
	for _, c := range ext {
		cName.Reset()
		cName.WriteString("bima:listener:")
		cName.WriteString(c)

		listeners = append(listeners, container.Get(cName.String()).(events.Listener))
	}

	ext = parsers.ParseMiddleware(workDir)
	hooks := make([]middlewares.Middleware, 0, len(ext))
	for _, c := range ext {
		cName.Reset()
		cName.WriteString("bima:middleware:")
		cName.WriteString(c)

		hooks = append(hooks, container.Get(cName.String()).(middlewares.Middleware))
	}

	ext = parsers.ParseLogger(workDir)
	extensions := make([]logrus.Hook, 0, len(ext))
	for _, c := range ext {
		cName.Reset()
		cName.WriteString("bima:logger:extension:")
		cName.WriteString(c)

		extensions = append(extensions, container.Get(cName.String()).(logrus.Hook))
	}

	ext = parsers.ParseRoute(workDir)
	handlers := make([]routes.Route, 0, len(ext))
	for _, c := range ext {
		cName.Reset()
		cName.WriteString("bima:route:")
		cName.WriteString(c)

		handlers = append(handlers, container.Get(cName.String()).(routes.Route))
	}

	ext = parsers.ParseRoute(workDir)
	storages := make([]drivers.Driver, 0, len(ext))
	for _, c := range ext {
		cName.Reset()
		cName.WriteString("bima:driver:")
		cName.WriteString(c)

		storages = append(storages, container.Get(cName.String()).(drivers.Driver))
	}

	container.GetBimaRouterMux().Register(handlers)
	container.GetBimaLoggerExtension().Register(extensions)
	container.GetBimaMiddlewareFactory().Register(hooks)
	container.GetBimaEventDispatcher().Register(listeners)
	container.GetBimaDriverFactory().Register(storages)
	container.GetBimaRouterGateway().Register(servers)

	util := color.New(color.FgGreen)
	util.Print("✓ ")
	fmt.Print("REST running on ")
	util.Println(env.HttpPort)
	fmt.Print(" with PID ")
	util.Println(os.Getpid())
	if env.Debug {
		util.Print("✓ ")
		fmt.Print("Api Doc ready on ")
		util.Println("/api/docs")
	}

	application := container.GetBimaApplication()
	loadInterface(container, application, *env)
	application.Run(servers)
}

func loadEnv(config *configs.Env, filePath string, ext string) {
	switch ext {
	case ".env":
		godotenv.Load()
		processDotEnv(config)
	case ".yaml":
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalln(err.Error())
		}

		err = yaml.Unmarshal(content, config)
		if err != nil {
			log.Fatalln(err.Error())
		}
	case ".json":
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalln(err.Error())
		}

		err = json.Unmarshal(content, config)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	if config.Secret == "" {
		hasher := sha256.New()
		hasher.Write([]byte(time.Now().Format(time.RFC3339)))

		config.Secret = base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	}
}

func processDotEnv(config *configs.Env) {
	config.Secret = os.Getenv("APP_SECRET")
	config.Debug, _ = strconv.ParseBool(os.Getenv("APP_DEBUG"))
	config.HttpPort, _ = strconv.Atoi(os.Getenv("APP_PORT"))
	config.RpcPort, _ = strconv.Atoi(os.Getenv("GRPC_PORT"))

	config.Service = os.Getenv("APP_NAME")
	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	config.Db = configs.Db{
		Host:     os.Getenv("DB_HOST"),
		Port:     dbPort,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Driver:   os.Getenv("DB_DRIVER"),
	}

	config.CacheLifetime, _ = strconv.Atoi(os.Getenv("CACHE_LIFETIME"))
}

func loadInterface(engine *engine.Container, application *interfaces.Factory, config configs.Env) {
	definition, err := engine.SafeGet("bima:interface:rest")
	rest, ok := definition.(*interfaces.Rest)
	if ok && err == nil {
		application.Add(rest)
	}

	if config.Db.Driver != "" {
		ctx := context.Background()
		options := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
		}

		var gRpcAddress strings.Builder
		gRpcAddress.WriteString("0.0.0.0:")
		gRpcAddress.WriteString(strconv.Itoa(config.RpcPort))

		gRpcClient, err := grpc.DialContext(ctx, gRpcAddress.String(), options...)
		if err != nil {
			log.Fatalf("Server is not ready. %v", err)
		}

		go func() {
			<-ctx.Done()
			if cerr := gRpcClient.Close(); cerr != nil {
				grpclog.Infof("Error closing connection to %s: %v", gRpcAddress, cerr)
			}
		}()

		rest.GRpcClient = gRpcClient
		application.Add(&interfaces.Database{})
		application.Add(&interfaces.GRpc{GRpcPort: config.RpcPort, Debug: config.Debug})
	}

	definition, err = engine.SafeGet("bima:interface:elasticsearch")
	if app, ok := definition.(*interfaces.Elasticsearch); ok && err == nil {
		application.Add(app)
	}

	definition, err = engine.SafeGet("bima:interface:consumer")
	if app, ok := definition.(*interfaces.Consumer); ok && err == nil {
		application.Add(app)
	}
}
