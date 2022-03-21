package cmd

import (
	"context"
	"github.com/nikitaksv/gendata/internal/service"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	"io"
	"os"
	"strings"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use: "gen",
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := service.NewService(zap.NewNop())
		rsp, err := srv.GenFile(context.Background(), &service.GenFileRequest{
			TmplFile: mustGetString(cmd.Flags(), "template-file"),
			DataFile: mustGetString(cmd.Flags(), "data-file"),
			Config: &service.Config{
				Lang:            mustGetString(cmd.Flags(), "lang"),
				DataFormat:      mustGetString(cmd.Flags(), "data-format"),
				RootClassName:   mustGetString(cmd.Flags(), "root-class-name"),
				PrefixClassName: mustGetString(cmd.Flags(), "prefix-class-name"),
				SuffixClassName: mustGetString(cmd.Flags(), "suffix-class-name"),
				SortProperties:  mustGetBool(cmd.Flags(), "sort-properties"),
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
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().StringP("template-file", "tmplf", "tmpl.txt", "")
	genCmd.Flags().StringP("data-file", "dataf", "data.json", "")
	genCmd.Flags().StringP("config-file", "cfgf", "config.json", "")
	genCmd.Flags().StringP("out", "o", ".", "")
	genCmd.Flags().StringP("lang", "l", "", "Programming language, supported: 'go','php'.")
	if err := genCmd.MarkFlagRequired("lang"); err != nil {
		panic(err)
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
