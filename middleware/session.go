// Copyright (c) 2017 Gorillalabs. All rights reserved.

package middleware

import (
	"fmt"
	"strings"

	"github.com/inosvaruag/go-powershell/utils"
	"github.com/pkg/errors"
)

// Default timeout for PS Session when idle.
const idleTimeoutMS = "60000"

type session struct {
	upstream Middleware
	name     string
}

func NewSession(upstream Middleware, config *SessionConfig) (Middleware, error) {
	asserted, ok := config.Credential.(credential)
	if ok {
		credentialParamValue, err := asserted.prepare(upstream)
		if err != nil {
			return nil, errors.Wrap(err, "Could not setup credentials")
		}

		config.Credential = credentialParamValue
	}

	psConfName := "$goConf" + utils.CreateRandomString(8)
	_, _, err := upstream.Execute(fmt.Sprintf("%s = New-PSSessionOption -IdleTimeout %s", psConfName, utils.QuoteArg(idleTimeoutMS)))
	if err != nil {
		return nil, errors.New("Could not convert password to secure string")
	}
	config.PSConfVar = psConfName

	name := "goSess" + utils.CreateRandomString(8)
	args := strings.Join(config.ToArgs(), " ")

	_, _, err = upstream.Execute(fmt.Sprintf("$%s = New-PSSession %s", name, args))
	if err != nil {
		return nil, errors.Wrap(err, "Could not create new PSSession")
	}

	return &session{upstream, name}, nil
}

func (s *session) Execute(cmd string) (string, string, error) {
	return s.upstream.Execute(fmt.Sprintf("Invoke-Command -Session $%s -Script {%s}", s.name, cmd))
}

func (s *session) Exit() {
	s.upstream.Execute(fmt.Sprintf("Disconnect-PSSession -Session $%s", s.name))
	s.upstream.Exit()
}
