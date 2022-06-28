package main

import (
	"context"
	"log"
	"time"

	"github.com/etcd-io/etcd/pkg/transport"
	"github.com/lxn/walk"
	"go.etcd.io/etcd/clientv3"
)

// ConnToEtcd connects to an ETCD database using TLS settings and returns the
// connection object
func connToEtcd(config Config) *clientv3.Client {
	tlsInfo := transport.TLSInfo{
		CertFile:      exePath + "\\" + config.Etcd.PeerCert,
		KeyFile:       exePath + "\\" + config.Etcd.PeerKey,
		TrustedCAFile: exePath + "\\" + config.Etcd.CertCa,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Etcd.Endpoints,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	return cli
}

// ReadFromEtcd reads all sub-prefixes from a given key and returns them in
// a map[string]string structure
func ReadFromEtcd(config Config, keyToRead string) map[string]string {
	cli := connToEtcd(config)
	defer cli.Close()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	gr, err := cli.Get(ctx, keyToRead, clientv3.WithPrefix())
	if err != nil {
		walk.MsgBox(mw, "Fatal Error", "Fatal: "+err.Error(), walk.MsgBoxIconError)
		log.Fatal(err)
	}

	answer := make(map[string]string)
	for i := range gr.Kvs {
		keyval := string(gr.Kvs[i].Key)
		answer[keyval] = string(gr.Kvs[i].Value)
	}

	return answer
}

// WriteToEtcd writes once to a given key in etcd
func WriteToEtcd(config Config, keyToWrite string, valueToWrite string) {
	cli := connToEtcd(config)
	defer cli.Close()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := cli.Put(ctx, keyToWrite, valueToWrite)
	if err != nil {
		log.Fatal(err)
	}
}

// DeleteFromEtcd delete the given key from etcd
func DeleteFromEtcd(config Config, keyToDelete string) {
	cli := connToEtcd(config)
	defer cli.Close()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := cli.Delete(ctx, keyToDelete)
	if err != nil {
		log.Fatal(err)
	}
}
