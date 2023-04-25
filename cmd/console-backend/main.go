package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Audience        string
	BindHost        string
	ConsoleEndpoint string
	ConsoleToken    string
	HookdEndpoint   string
	HookdPSK        string
	Kubeconfig      string
	LogLevel        string
	Port            string
}

var cfg = &Config{}

func init() {
	flag.StringVar(&cfg.Audience, "audience", os.Getenv("IAP_AUDIENCE"), "IAP audience")
	flag.StringVar(&cfg.BindHost, "bind-host", os.Getenv("BIND_HOST"), "Bind host")
	flag.StringVar(&cfg.ConsoleEndpoint, "console-endpoint", envOrDefault("CONSOLE_ENDPOINT", "http://console.local.nais.io/query"), "Console endpoint")
	flag.StringVar(&cfg.ConsoleToken, "console-token", envOrDefault("CONSOLE_TOKEN", "secret"), "Console Token")
	flag.StringVar(&cfg.HookdEndpoint, "hookd-endpoint", envOrDefault("HOOKD_ENDPOINT", "http://hookd.local.nais.io"), "Hookd endpoint")
	flag.StringVar(&cfg.HookdPSK, "hookd-psk", envOrDefault("HOOKD_PSK", "secret-frontend-psk"), "Hookd PSK")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "which log level to output")
	flag.StringVar(&cfg.Port, "port", envOrDefault("PORT", "8080"), "Port to listen on")
	flag.StringVar(&cfg.Kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "kubeconfigpath")
}

func main() {
	flag.Parse()
	log := newLogger()
	ctx := context.Background()

	log.WithField("path", cfg.Kubeconfig).Debug("reading kubeconfig")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	kubeConfig.
		log.Debug("starting k8s client")
	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.WithError(err).Fatal("setting up k8s client")
	}

	ns, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	fmt.Println(ns)
	/*
	   	graphConfig := graph.Config{
	   		Resolvers: &graph.Resolver{
	   			Hookd:   hookd.New(cfg.HookdPSK, cfg.HookdEndpoint),
	   			Console: console.New(cfg.ConsoleToken, cfg.ConsoleEndpoint),
	   		},
	   	}

	   srv := handler.NewDefaultServer(graph.NewExecutableSchema(graphConfig))

	   http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	   http.Handle("/query", auth.InsecureValidateMW(srv))

	   log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.Port)
	   log.Fatal(http.ListenAndServe(cfg.BindHost+":"+cfg.Port, nil))
	*/
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func newLogger() *logrus.Logger {
	log := logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(l)
	return log
}
