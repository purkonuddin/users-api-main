package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/jackc/pgx"
)

func (c *InitAPI) initDb() {
	dbHost := "127.0.0.1"
	dbPass := "upgrading1"
	dbName := "postgres"
	dbPort := "5432"
	dbUser := "postgres"

	port, err := strconv.Atoi(dbPort)
	if err != nil {
		log.Println(err)
		return
	}

	dbConfig := &pgx.ConnConfig{
		Port:     uint16(port),
		Host:     dbHost,
		User:     dbUser,
		Password: dbPass,
		Database: dbName,
	}

	connection := pgx.ConnPoolConfig{
		ConnConfig:     *dbConfig,
		MaxConnections: 5,
	}

	c.Db, err = pgx.NewConnPool(connection)
	if err != nil {
		log.Println(err)
		return
	}
}

/*
func (c *InitAPI) HandleGetUserById(w http.ResponseWriter, r *http.Request) {
	resp := "asdf"
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-conver-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
*/

/*
	context untuk membatasi waktu, timeout:
	WithCancel cancel setelah semuanya di...
	WithTimeout ...
*/
func (c *InitAPI) HandleListUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // dilakukan ketika semuanya fungsi telah berakhir

	var p GetUsers
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "failed-to-convert-json", http.StatusBadRequest)
		return
	}

	resp, err := c.ListUser(ctx, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-conver-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *InitAPI) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p User
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "failed-to-convert-json", http.StatusBadRequest)
		return
	}

	roleid := r.Header.Get("ROLE-ID")
	resp, err := c.CreateUser(ctx, &p, roleid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-conver-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hola... web server 127.0.0.1:3000!")
}

func (c *InitAPI) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p User
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "failed-to-convert-json", http.StatusBadRequest)
		return
	}

	// id := r.Header.Get("USER-ID")
	id := r.URL.Query().Get("id")

	resp, err := c.UpdateUser(ctx, &p, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-conver-json", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}

func (c *InitAPI) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// id := r.URL.Query().Get("id")
	id := r.FormValue("id")
	resp, err := c.DeleteUser(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-conver-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}

// HandleUploadPhoto
/*
cannot convert 101933 to Text
Error: 101933 is greater than maximum value for QChar
open asset/beeflasagna.png: The system cannot find the path specified.
*/
func (c *InitAPI) HandleUploadPhoto(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.ParseMultipartForm(1024); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

	id := r.FormValue("userId")
	resp, err := c.InsertProfilePhoto(ctx, &FileItem{
		File:     file,
		UserId:   id,
		Filename: header.Filename,
		FileSize: header.Size,
		FileType: header.Header["Content-Type"][0],
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed-convert-json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// HandleGetProfilePhoto
func (c *InitAPI) HandleGetProfilePhoto(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["userid"]

	img, fileType, err := c.GetProfilePhoto(r.Context(), &GetFile{
		UserId: id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", fileType) // <- set content type header
	io.Copy(w, img)
	/*
		data, err := json.Marshal(img)
		if err != nil {
			http.Error(w, "failed-convert-json", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	*/
}

func StartHttp() http.Handler {
	api := createAPI()
	api.initDb()

	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/api/user/list", api.HandleListUser).Methods("GET")
	// r.HandleFunc("/api/user/list/{userid}", api.HandleGetUserById).Methods("GET")
	r.HandleFunc("/api/user/create", api.HandleCreateUser).Methods("POST")
	r.HandleFunc("/api/user/update", api.HandleUpdateUser).Methods("PATCH")
	r.HandleFunc("/api/user/delete", api.HandleDeleteUser).Methods("DELETE")
	r.HandleFunc("/api/user/photo", api.HandleUploadPhoto).Methods("POST")
	r.HandleFunc("/api/user/photo/{userid}", api.HandleGetProfilePhoto).Methods("GET")

	return r
}
