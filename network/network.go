/*
Copyright (C) 2016-2017 dapperdox.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/

// Package network creates connections with or without TLS.
package network

import (
	"crypto/tls"
	"net"

	"github.com/spf13/viper"

	"github.com/kenjones-cisco/dapperdox/config"
)

// NewListener creates a new network Listener
func NewListener() (net.Listener, error) {
	log().Infof("listening on %s for unsecured connections", viper.GetString(config.BindAddr))
	return newListener()
}

// NewSecuredListener creates a secure network Listener
func NewSecuredListener() (net.Listener, error) {
	crt, err := tls.LoadX509KeyPair(viper.GetString(config.TLSCert), viper.GetString(config.TLSKey))
	if err != nil {
		return nil, err
	}

	// Inspired by https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
	tlscfg := &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		// https://github.com/golang/go/tree/master/src/crypto/elliptic
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		// Use modern tls mode https://wiki.mozilla.org/Security/Server_Side_TLS#Modern_compatibility
		NextProtos: []string{"h2", "http/1.1"},
		// https://www.owasp.org/index.php/Transport_Layer_Protection_Cheat_Sheet#Rule_-_Only_Support_Strong_Protocols
		MinVersion: tls.VersionTLS12,
		// These ciphersuites support Forward Secrecy: https://en.wikipedia.org/wiki/Forward_secrecy
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		Certificates: []tls.Certificate{crt},
	}

	l, err := newListener()
	if err != nil {
		return nil, err
	}
	log().Infof("listening on %s for SECURED connections", viper.GetString(config.BindAddr))
	return tls.NewListener(l, tlscfg), nil
}

func newListener() (net.Listener, error) {
	return net.Listen("tcp", viper.GetString(config.BindAddr))
}
