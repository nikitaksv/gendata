package cmd

import (
	"context"
	"io"
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
			if err = os.WriteFile(outPath+file.FileName, bs, 0666); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	genCmd.Flags().StringP("tmplFile", "t", "tmpl.txt", "")
	genCmd.Flags().StringP("dataFile", "d", "data.json", "")
	genCmd.Flags().StringP("configFile", "c", "config.json", "")
	genCmd.Flags().StringP("out", "o", ".", "")
	genCmd.Flags().StringP("lang", "l", "", "Programming language, supported: 'go','php'.")
	genCmd.Flags().StringP("dataFormat", "", "json", "")
	genCmd.Flags().StringP("rootClassName", "", "RootClass", "")
	genCmd.Flags().StringP("prefixClassName", "", "", "")
	genCmd.Flags().StringP("suffixClassName", "", "", "")
	genCmd.Flags().BoolP("sort", "", false, "Sort object properties")

	if err := genCmd.MarkFlagRequired("lang"); err != nil {
		panic(err)
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
