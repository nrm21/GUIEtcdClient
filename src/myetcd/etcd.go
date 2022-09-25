package myetcd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/etcd-io/etcd/pkg/transport"
	"go.etcd.io/etcd/clientv3"
)

const dbTimeoutTime = 5

// ConnToEtcd connects to an ETCD database using TLS settings and returns the connection object
func connToEtcd(certPath *string, endpoints *[]string) *clientv3.Client {
	tlsInfo := transport.TLSInfo{
		CertFile:      *certPath + "\\peer.crt",
		KeyFile:       *certPath + "\\peer.key",
		TrustedCAFile: *certPath + "\\ca.crt",
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   *endpoints,
		DialTimeout: dbTimeoutTime * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	return cli
}

// ReadFromEtcd reads all sub-prefixes from a given key and returns them in
// a (string, byte array) map structure
func ReadFromEtcd(certPath *string, endpoints *[]string, keyToRead string) (map[string][]byte, error) {
	cli := connToEtcd(certPath, endpoints)
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeoutTime*time.Second)
	defer cancel()

	response, err := cli.Get(ctx, keyToRead, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// convert their weird KV struct into generic map[string]string before return
	answer := make(map[string][]byte)
	for i := range response.Kvs {
		keyval := string(response.Kvs[i].Key)
		answer[keyval] = response.Kvs[i].Value
	}

	return answer, nil
}

// WatchReadFromEtcd watches all sub-prefixes from a given key and returns any
// changes to them in a (string, byte array) map structure, this fuction loops
// forever unless broken from explicitly
func WatchReadFromEtcd(certPath *string, endpoints *[]string, keyToRead string, watchedChangeCh chan map[string][]byte, closeWacher chan bool) {
	fmt.Printf("Now watching %s!\n", keyToRead)

	cli := connToEtcd(certPath, endpoints)
	defer cli.Close()
	modifiedKv := make(map[string][]byte)

	watchChan := cli.Watch(context.Background(), keyToRead, clientv3.WithPrefix())
	for {
		select {
		case <-watchChan:
			wc := <-watchChan
			for _, ev := range wc.Events {
				// ev.Type,  ev.Kv.Key,  ev.Kv.Value
				keyval := string(ev.Kv.Key)
				modifiedKv[keyval] = ev.Kv.Value
				watchedChangeCh <- modifiedKv
			}
		default:
			// check to see if we should close the watcher to start one on a
			// different key.  If not we should just wait a short time then
			// return control to the main for loop to check for more watch
			// events in the code above.
			select {
			case _, ok := <-closeWacher:
				if ok {
					fmt.Printf("Killing watcher for %s!\n", keyToRead)
					return
				}
			case <-time.After(250 * time.Millisecond):
				continue
			}
		}
	}
}

// WriteToEtcd writes once to a given key in etcd
func WriteToEtcd(certPath *string, endpoints *[]string, keyToWrite string, valueToWrite string) {
	cli := connToEtcd(certPath, endpoints)
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeoutTime*time.Second)
	defer cancel()

	_, err := cli.Put(ctx, keyToWrite, valueToWrite)
	if err != nil {
		log.Fatal(err)
	}
}

// Deletes the given key from etcd and returns the amount deleted
func DeleteFromEtcd(certPath *string, endpoints *[]string, keyToDelete string) int64 {
	var response *clientv3.DeleteResponse
	cli := connToEtcd(certPath, endpoints)
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeoutTime*time.Second)
	defer cancel()

	response, err := cli.Delete(ctx, keyToDelete)
	if err != nil {
		log.Fatal(err)
	}

	return response.Deleted
}
