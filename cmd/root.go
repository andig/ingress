package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	cfgLogLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ingress",
	Short: "ingress data mapper daemon",
	Run:   run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config", "",
		"config file (default is $HOME/ingress.yaml)",
	)

	rootCmd.PersistentFlags().BoolP(
		"dump", "d",
		false,
		"Dump parsed config",
	)
	rootCmd.PersistentFlags().BoolP(
		"capabilities", "i",
		false,
		"Display available capabilities",
	)
	rootCmd.PersistentFlags().Bool(
		"diagnose",
		false,
		"Memory diagnostics",
	)
	rootCmd.PersistentFlags().StringVarP(
		&cfgLogLevel,
		"log", "l",
		"debug",
		"Log level (error, info, debug, trace)",
	)
	rootCmd.PersistentFlags().BoolP(
		"test", "t",
		false,
		"Inject test data",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".ingress" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")    // optionally look for config in the working directory
		viper.AddConfigPath("/etc") // path to look for the config file in

		viper.SetConfigName("ingress")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	cfgFile = viper.ConfigFileUsed()
	fmt.Println("Using config file:", cfgFile)
}
