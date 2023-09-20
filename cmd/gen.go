package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nikitaksv/gendata/pkg/gen"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use: "gen",
	RunE: func(cmd *cobra.Command, args []string) error {
		g := gen.NewGen()

		tmplDirPath := mustGetString(cmd.Flags(), "tmplDir")
		entries, err := os.ReadDir(tmplDirPath)
		if err != nil {
			return err
		}

		tmplFiles := make([]*gen.File, 0, 3)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) == ".tmpl" {
				fullPath := filepath.Join(tmplDirPath, entry.Name())
				f, err := os.Open(fullPath)
				if err != nil {
					return errors.WithMessagef(err, "can't open file \"%s\"", fullPath)
				}
				tmplFiles = append(tmplFiles, &gen.File{
					Name: entry.Name(),
					Body: f,
				})
			}
		}

		dataFilePath := mustGetString(cmd.Flags(), "dataFile")
		dataFileBody, err := os.Open(dataFilePath)
		if err != nil {
			return errors.WithMessagef(err, "can't open data file \"%s\"", dataFilePath)
		}
		dataFile := &gen.File{
			Name: filepath.Base(dataFilePath),
			Body: dataFileBody,
		}

		generatedFiles, err := g.Gen(context.Background(), &gen.Params{
			RootClassName:   mustGetString(cmd.Flags(), "rootClassName"),
			PrefixClassName: mustGetString(cmd.Flags(), "prefixClassName"),
			SuffixClassName: mustGetString(cmd.Flags(), "suffixClassName"),
			SortProperties:  mustGetBool(cmd.Flags(), "sort"),
			Templates:       tmplFiles,
			Data:            dataFile,
		})
		if err != nil {
			return err
		}

		duplNames := map[string]int{}
		for _, file := range generatedFiles.RenderedFiles {
			if count, ok := duplNames[file.Name]; ok {
				duplNames[file.Name] = count + 1
			} else {
				duplNames[file.Name] = 0
			}
		}

		outPath := mustGetString(cmd.Flags(), "out")
		for _, file := range generatedFiles.RenderedFiles {
			count := duplNames[file.Name]
			name := file.Name
			if count > 0 {
				name = strconv.Itoa(count) + "_" + name
			}
			path := filepath.Join(outPath, name)
			bs, err := io.ReadAll(file.Body)
			if err != nil {
				return err
			}
			if err = os.WriteFile(path, bs, 0600); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	genCmd.Flags().StringP("tmplDir", "t", "", "Path to directory with template files")
	genCmd.Flags().StringP("dataFile", "d", "", "Path to data file")
	genCmd.Flags().StringP("out", "o", ".", "Path to output files directory")
	genCmd.Flags().StringP("rootClassName", "", "", "Name for root (first) object in data")
	genCmd.Flags().StringP("prefixClassName", "", "", "Prefix name class")
	genCmd.Flags().StringP("suffixClassName", "", "", "Suffix name class")
	genCmd.Flags().BoolP("sort", "", false, "Sort data objects properties")

	if err := genCmd.MarkFlagRequired("tmplDir"); err != nil {
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
