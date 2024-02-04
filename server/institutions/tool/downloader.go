package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"sort"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/teller"
	"github.com/plaid/plaid-go/v14/plaid"
)

var (
	plaidClientId = flag.String("plaid-client-id", "", "Plaid API Client ID")
	plaidSecret   = flag.String("plaid-secret", "", "Plaid API Secret")
	plaidOutput   = flag.String("plaid-output", "", "Output JSON file for Plaid")
	tellerOutput  = flag.String("teller-output", "", "Output JSON file for Teller")
)

func main() {
	flag.Parse()

	if *plaidClientId == "" || *plaidSecret == "" {
		log.Fatal("plaid client ID and plaid secret must be provided")
		return
	}

	conf := plaid.NewConfiguration()
	conf.UseEnvironment(plaid.Sandbox)
	conf.AddDefaultHeader("PLAID-CLIENT-ID", *plaidClientId)
	conf.AddDefaultHeader("PLAID-SECRET", *plaidSecret)

	plaidClient := plaid.NewAPIClient(conf)
	tellerClient, _ := teller.NewClient(nil, config.Teller{})

	if *tellerOutput != "" {
		log.Println("downloading institutions from teller")
		tellerInstitutions, err := getTellerInstitutions(tellerClient)
		if err != nil {
			log.Fatalf("failed to retrieve teller institutions: %+v", err)
			return
		}

		file, err := os.OpenFile(*tellerOutput, os.O_RDWR, 0755)
		if err != nil {
			log.Fatalf("failed to open teller output file", err)
			return
		}

		sort.Slice(tellerInstitutions, func(i, j int) bool {
			return tellerInstitutions[i].Id < tellerInstitutions[j].Id
		})

		if err := json.NewEncoder(file).Encode(tellerInstitutions); err != nil {
			log.Fatalf("failed to encode teller institutions to json", err)
			return
		}

		file.Sync()
		file.Close()
	}

	if *tellerOutput != "" {
		log.Println("downloading institutions from plaid")
		tellerInstitutions, err := getTellerInstitutions(tellerClient)
		if err != nil {
			log.Fatalf("failed to retrieve teller institutions: %+v", err)
			return
		}

		file, err := os.OpenFile(*tellerOutput, os.O_RDWR, 0755)
		if err != nil {
			log.Fatalf("failed to open teller output file", err)
			return
		}

		sort.Slice(tellerInstitutions, func(i, j int) bool {
			return tellerInstitutions[i].Id < tellerInstitutions[j].Id
		})

		if err := json.NewEncoder(file).Encode(tellerInstitutions); err != nil {
			log.Fatalf("failed to encode teller institutions to json", err)
			return
		}

		file.Sync()
		file.Close()
	}

}

func getTellerInstitutions(client teller.Client) ([]teller.Institution, error) {
	institutions, err := client.GetInstitutions(context.Background())
	return institutions, err
}

func getPlaidInstitutions(client *plaid.APIClient) ([]plaid.Institution, error) {
	institutions := make([]plaid.Institution, 0)
	pageSize := 500
	for {
		request := client.PlaidApi.
			InstitutionsGet(context.Background()).
			InstitutionsGetRequest(plaid.InstitutionsGetRequest{
				Count:  int32(pageSize),
				Offset: int32(len(institutions)),
				Options: &plaid.InstitutionsGetRequestOptions{
					Products: []plaid.Products{
						plaid.PRODUCTS_TRANSACTIONS,
					},
				},
			})

		result, _, err := request.Execute()
		if err != nil {
			return nil, err
		}

		items := result.GetInstitutions()
		institutions = append(institutions, items...)
		if len(items) < pageSize {
			break
		}
	}

	return institutions, nil
}
