package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/nikitaksv/gendata/pkg/service"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use: "gen",
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := service.NewService(zap.NewNop())
		rsp, err := srv.GenFile(context.Background(), &service.GenFileRequest{
			TmplFile:   mustGetString(cmd.Flags(), "tmplFile"),
			DataFile:   mustGetString(cmd.Flags(), "dataFile"),
			ConfigFile: mustGetString(cmd.Flags(), "configFile"),
			Config: &service.Config{
				Lang:            mustGetString(cmd.Flags(), "lang"),
				DataFormat:      mustGetString(cmd.Flags(), "dataFormat"),
				RootClassName:   mustGetString(cmd.Flags(), "rootClassName"),
				PrefixClassName: mustGetString(cmd.Flags(), "prefixClassName"),
				SuffixClassName: mustGetString(cmd.Flags(), "suffixClassName"),
				SortProperties:  mustGetBool(cmd.Flags(), "sort"),
			},
		})
		if err != nil {
			return err
		}

		outPath := mustGetString(cmd.Flags(), "out")
		if !strings.HasSuffix(outPath, string(os.PathSeparator)) {
			outPath += string(os.PathSeparator)
		}

		for _, file := range rsp.RenderedFiles {
			bs, err := io.ReadAll(file.Content)
			if err != nil {
				return err
			}
			if err = os.WriteFile(outPath+file.FileName, bs, 0600); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	genCmd.Flags().StringP("tmplFile", "t", "", "Path to template file")
	genCmd.Flags().StringP("dataFile", "d", "", "Path to data file")
	genCmd.Flags().StringP("configFile", "c", "", "Path to config.[json,xml,yaml,yml] file")
	genCmd.Flags().StringP("out", "o", ".", "Path to output files directory")
	genCmd.Flags().StringP("lang", "l", "", "Programming language, supported: 'go','php'.")
	genCmd.Flags().StringP("dataFormat", "", "", "Set manual data format [json] or auto-detect by file extension!")
	genCmd.Flags().StringP("rootClassName", "", "RootClass", "Name for root (first) object in data")
	genCmd.Flags().StringP("prefixClassName", "", "", "Prefix name class")
	genCmd.Flags().StringP("suffixClassName", "", "", "Suffix name class")
	genCmd.Flags().BoolP("sort", "", false, "Sort data objects properties")

	if err := genCmd.MarkFlagRequired("lang"); err != nil {
		log.Fatal(err)
	}
	if err := genCmd.MarkFlagRequired("tmplFile"); err != nil {
		log.Fatal(err)
	}
	if err := genCmd.MarkFlagRequired("dataFile"); err != nil {
		log.Fatal(err)
	}
}

func Execute() {
	err := genCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
func mustGetBool(f *flag.FlagSet, name string) bool {
	v, err := f.GetBool(name)
	if err != nil {
		panic(err)
	}
	return v
}

func mustGetString(f *flag.FlagSet, name string) string {
	v, err := f.GetString(name)
	if err != nil {
		panic(err)
	}
	return v
}
