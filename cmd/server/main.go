package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/anxiousmodernman/gnolly/models"
	"github.com/boltdb/bolt"
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

func main() {

	ctx := context.Background()
	database, err := bolt.Open("gnolly.db", 0600, nil)
	defer database.Close()

	if err != nil {
		panic(err)
	}

	_ = database.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("main"))
		return nil
	})

	svc := gnolly{database}

	getHandler := httptransport.NewServer(
		ctx,
		makeGetEndpoint(&svc),
		decodeGetRequest,
		encodeResponse,
	)

	putHandler := httptransport.NewServer(
		ctx,
		makePutEndpoint(&svc),
		decodePutRequest,
		encodeResponse,
	)

	http.Handle("/get", getHandler)
	http.Handle("/put", putHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func decodeGetRequest(r *http.Request) (interface{}, error) {
	var request getRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodePutRequest(r *http.Request) (interface{}, error) {
	var request putRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// Gnolly implements basic gets and puts to the database
type Gnolly interface {
	Put(models.Item) error
	Get(string) (models.Item, error)
}

type gnolly struct {
	db *bolt.DB
}

// Put implements put
func (g *gnolly) Put(item models.Item) error {

	err := g.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("main"))
		err := b.Put([]byte(item.Key), []byte(item.Value))
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// Get implements get
func (g *gnolly) Get(key string) (models.Item, error) {
	var value []byte
	err := g.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("main"))
		value = b.Get([]byte(key))
		return nil
	})
	return models.Item{Key: key, Value: string(value)}, err
}

func makeGetEndpoint(svc *gnolly) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getRequest)
		item, err := svc.Get(req.Key)
		if err != nil {
			return getResponse{"", ""}, nil
		}
		return getResponse{item.Key, item.Value}, nil
	}
}

func makePutEndpoint(svc *gnolly) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(putRequest)
		item, err := models.NewItem(req.Key, req.Value)
		if err != nil {
			return putResponse{err.Error()}, nil
		}
		err = svc.Put(item)
		if err != nil {
			return putResponse{err.Error()}, nil
		}
		return putResponse{""}, nil
	}
}

type putRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type putResponse struct {
	Err string `json:"err"`
}

type getRequest struct {
	Key string `json:"key"`
}

type getResponse struct {
	K string `json:"key"`
	V string `json:"value"`
}
