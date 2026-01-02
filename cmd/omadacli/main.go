package main

import (
	"context"
	"crypto/tls"
	"flag"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/rmrobinson/omada"
	"github.com/rmrobinson/omada/api"
	"go.uber.org/zap"
)

func main() {
	controllerURL := flag.String("url", "The URL the controller can be reached at", "")
	cid := flag.String("cid", "The Omada controller ID", "")
	clientID := flag.String("clientID", "The Omada OAuth token client_id", "")
	clientSecret := flag.String("clientSecret", "The Omada OAuth token client_secret", "")

	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // self-hosted Omada SDN controllers use self-signed certs. If you have a proper cert, don't use this.
			},
		},
	}

	omadaClient := omada.NewClient(logger, *controllerURL, *cid, *clientID, *clientSecret, httpClient)

	sites, err := omadaClient.GetSites(context.Background())
	if err != nil {
		logger.Error("unable to obtain sites", zap.Error(err))
	}
	for _, site := range sites {
		logger.Debug("site", zap.String("id", site.ID), zap.String("name", site.Name))

		trueArg := "true"
		req := &api.GetGridActiveClientsParams{
			Page:            1,
			PageSize:        50,
			FiltersWireless: &trueArg,
			SortsMac:        &trueArg,
		}
		resp, err := omadaClient.GetGridActiveClientsWithResponse(context.Background(), *cid, site.ID, req)
		if err != nil {
			logger.Error("unable to get clients for site", zap.Error(err), zap.String("site_id", site.ID))
			continue
		}

		for _, activeClient := range *resp.JSON200.Result.Data {
			spew.Dump(activeClient)
		}
	}
}
