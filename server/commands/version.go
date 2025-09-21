package commands

import (
	"fmt"
	"runtime"
	"strings"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/icons"
	"github.com/monetr/monetr/server/ui"
	"github.com/spf13/cobra"
	"golang.org/x/sys/cpu"
)

func VersionCommand(parent *cobra.Command) {
	var arguments struct {
		detailed bool
	}

	command := &cobra.Command{
		Use:   "version",
		Short: "Print the version of monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !arguments.detailed {
				fmt.Println(build.Release)
				return nil
			}

			detailedString := "" +
				"Version:         %s\n" +
				"Revision:        %s\n" +
				"Build time:      %s\n" +
				"Build host:      %s\n" +
				"Build type:      %s\n" +
				"Embedded UI:     %t\n" +
				"Embedded Icons:  %t\n" +
				"  Icon Packs:    %s\n" +
				"Locales:         %d\n" +
				"Currencies:      %d\n" +
				"Architecture:    %s\n" +
				"OS:              %s\n" +
				"SIMD:            %s\n" +
				"Compiler:        %s\n" +
				"Go Version:      %s\n"

			locales := locale.GetInstalledLocales()
			currencies := locale.GetInstalledCurrencies()

			iconsEnabled := icons.GetIconsEnabled()
			iconPacks := "<not enabled>"
			if iconsEnabled {
				indexes := icons.GetIconIndexes()
				if len(indexes) == 0 {
					iconPacks = "<none enabled>"
				} else {
					iconPacks = strings.Join(indexes, ", ")
				}
			}

			simd := "N/A"
			if cpu.X86.HasAVX512F {
				simd = "AVX512"
			} else if cpu.X86.HasAVX {
				simd = "AVX"
			}

			fmt.Printf(
				detailedString,
				build.Release,
				build.Revision,
				build.BuildTime,
				build.BuildHost,
				build.BuildType,
				ui.EmbeddedUI,
				iconsEnabled,
				iconPacks,
				len(locales),
				len(currencies),
				runtime.GOARCH,
				runtime.GOOS,
				simd,
				runtime.Compiler,
				runtime.Version(),
			)

			return nil
		},
	}

	command.PersistentFlags().BoolVarP(&arguments.detailed, "detailed", "d", false, "Print detailed version information.")

	parent.AddCommand(command)
}
