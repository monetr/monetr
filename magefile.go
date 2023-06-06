//go:build mage

package main

import (
	"fmt"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

var (
	progress *mpb.Progress
)

func init() {
	progress = mpb.New(mpb.WithWidth(64))
}

var (
	goSrcDirectory = "./pkg"
	nodeModules    = "./node_modules"
)

type Go mg.Namespace

func (Go) Dependencies() error {
	bar := progress.AddBar(100,
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name("go:dependencies"),
			// decor.DSyncWidth bit enables column width synchronization
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 30
				decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "done",
			),
		),
	)
	defer bar.SetCurrent(100)
	outdated, err := target.Path("./vendor", "go.mod", "go.sum")
	if err != nil {
		return err
	}
	if !outdated {
		return nil
	}
	fmt.Println("installing go dependencies")
	_, err = sh.Output("go", "mod", "vendor", "-o", "./vendor")
	return err
}

type UI mg.Namespace

func (UI) Dependencies() error {
	bar := progress.AddBar(100,
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name("ui:dependencies"),
			// decor.DSyncWidth bit enables column width synchronization
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 30
				decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "done",
			),
		),
	)
	defer bar.SetCurrent(100)
	outdated, err := target.Path(nodeModules, "package.json", "yarn.lock")
	if err != nil {
		return err
	}
	if !outdated {
		fmt.Println("no need to install ui deps")
		return nil
	}
	fmt.Println("installing node dependencies")

	switch runtime.GOOS {
	case "darwin":
		if _, err := sh.Output("yarn", "install", "--ignore-platform"); err != nil {
			return err
		}
	default:
		if _, err := sh.Output("yarn", "install"); err != nil {
			return err
		}
	}
	if runtime.GOOS != "windows" {
		_, err := sh.Output("touch", "-a", "-m", nodeModules)
		return err
	}
	return nil
}

func (UI) Build() error {
	bar := progress.AddBar(100,
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name("ui:build"),
			// decor.DSyncWidth bit enables column width synchronization
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 30
				decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "done",
			),
		),
	)
	defer bar.SetCurrent(100)
	mg.Deps(UI.Dependencies)
	outdated, err := target.Glob("pkg/ui/static/index.html", "./ui/**/**")
	if err != nil {
		return err
	}
	if !outdated {
		fmt.Println("no need to rebuild ui")
		return nil
	}
	fmt.Println("building frontend")
	if _, err := sh.Output("yarn", "build:production"); err != nil {
		return err
	}
	return nil
}

func Build() error {
	bar := progress.AddBar(100,
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name("build"),
			// decor.DSyncWidth bit enables column width synchronization
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 30
				decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "done",
			),
		),
	)
	defer bar.SetCurrent(100)
	mg.Deps(Go.Dependencies, UI.Dependencies, UI.Build)
	fmt.Println("building monetr")
	_, err := sh.Output("go", "build", "-mod=vendor", "-o", "build/monetr", "github.com/monetr/monetr/pkg/cmd")
	return err
}
