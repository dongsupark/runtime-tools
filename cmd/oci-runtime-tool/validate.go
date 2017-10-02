package main

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	rfc2119 "github.com/opencontainers/runtime-tools/error"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validate"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var bundleValidateFlags = []cli.Flag{
	cli.StringFlag{Name: "path", Value: ".", Usage: "path to a bundle"},
	cli.StringFlag{Name: "platform", Value: "linux", Usage: "platform of the target bundle (linux, windows, solaris)"},
}

var bundleValidateCommand = cli.Command{
	Name:   "validate",
	Usage:  "validate an OCI bundle",
	Flags:  bundleValidateFlags,
	Before: before,
	Action: func(context *cli.Context) error {
		hostSpecific := context.GlobalBool("host-specific")
		complianceLevelString := context.GlobalString("compliance-level")
		complianceLevel, err := rfc2119.ParseLevel(complianceLevelString)
		if err != nil {
			complianceLevel = rfc2119.Must
			logrus.Warningf("%s, using 'MUST' by default.", err.Error())
		}
		inputPath := context.String("path")
		platform := context.String("platform")
		v, err := validate.NewValidatorFromPath(inputPath, hostSpecific, platform)
		if err != nil {
			return err
		}

		if err := v.CheckAll(); err != nil {
			merr, ok := err.(*multierror.Error)
			if !ok {
				return err
			}
			var validationErrors error
			for _, err = range merr.Errors {
				e, ok := err.(*specerror.Error)
				if ok && e.Err.Level < complianceLevel {
					logrus.Warn(e)
					continue
				}
				validationErrors = multierror.Append(validationErrors, err)
			}
			return validationErrors
		}
		fmt.Println("Bundle validation succeeded.")
		return nil
	},
}
