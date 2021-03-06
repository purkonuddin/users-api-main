package restapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func (c *InitAPI) GetProfilePhotoById(id string) (string, string, error) {
	var filename, fileType string
	err := c.Db.QueryRow(`SELECT filename, file_type FROM photo_users WHERE user_id = $1`, id).Scan(&filename, &fileType)
	if err != nil {
		log.Println(err)
		return "", "", err
	}
	return filename, fileType, nil
}

// GetProfilePhoto
func (c *InitAPI) GetProfilePhoto(ctx context.Context, req *GetFile) (io.Reader, string, error) {
	filename, fileType, err := c.GetProfilePhotoById(req.UserId)
	if err != nil {
		return nil, "", nil
	}

	url := fmt.Sprintf("../asset/%s", filename)
	file, err := os.Open(url)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}

	return file, fileType, nil
}

// InsertProfilePhoto
func (c *InitAPI) InsertProfilePhoto(ctx context.Context, req *FileItem) (*UserId, error) {
	var profileId string
	err := c.Db.QueryRow(`INSERT INTO photo_users (user_id, filename, file_type, size) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.UserId, req.Filename, req.FileType, req.FileSize).Scan(&profileId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	filename := fmt.Sprintf("../asset/%s", req.Filename)

	// fileLocation := filepath.Join(dir, "files", req.Filename)
	// targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
	// if err != nil {
	// 	log.Println(err)
	// 	return nil, err
	// }
	// defer targetFile.Close()

	file, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer file.Close()
	// _, err := io.Copy(file, req.File)
	// if err != nil {
	// 	log.Println(err)
	// 	return nil, err
	// }
	if _, err := io.Copy(file, req.File); err != nil {
		log.Println(err)
		return nil, err
	}

	return &UserId{
		Id: profileId,
	}, nil
}

// GetCustomerById
func (c *InitAPI) GetCustomerById(id string) bool {
	var userId string
	// err := c.Db.QueryRow(context.Background(), `SELECT username FROM users WHERE id = $1`, id).Scan(&userId)
	// if err != nil {
	// 	return false
	// }

	return userId != ""
}

// UpdateUser
func (c *InitAPI) UpdateUser(ctx context.Context, req *User, id string) (*ReturnActions, error) {
	updateQry := `update "users" set "email" = $1, "status" = $2, "role_id" = $3, "updated_at" = $4 where "id" = $5`

	_, err := c.Db.Exec(updateQry, req.Email, req.Status, req.RoleId, time.Now(), id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &ReturnActions{
		Object:   "users",
		Actions:  "UpdateUser",
		Id:       id,
		Messages: "Success",
	}, nil
}

// DeleteUser
func (c *InitAPI) DeleteUser(ctx context.Context, id string) (*ReturnActions, error) {
	deleteQry := `delete from "users" where id = $1`

	result, err := c.Db.Exec(deleteQry, id)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(result)

	// data, err := json.Marshal(result)
	// if err != nil {
	// 	return nil, err
	// }

	// return data, nil

	return &ReturnActions{
		Object:   "users",
		Actions:  "Delete",
		Id:       id,
		Messages: "Successs",
	}, nil
}

// ListUser
func (c *InitAPI) ListUser(ctx context.Context, req *GetUsers) (*GetUsers, error) {
	limit := 10

	if req.Limit != 0 {
		limit = int(req.Limit)
	}

	rows, err := c.Db.Query(`
		SELECT id, 
			username, 
			email,
			status, 
			role_id,
			created_at,
			updated_at
		FROM users LIMIT $1
	`, limit)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	var items []*User
	for rows.Next() {
		var item User
		// var updateTime sql.NullString
		// var status string
		err = rows.Scan(&item.Id,
			&item.Username,
			&item.Email,
			&item.Status,
			&item.RoleId,
			&item.CreatedAt,
			&item.UpdatedAt,
		)

		if err != nil {
			log.Println(err)
			return nil, err
		}

		// item.UpdatedAt = updateTime.String

		items = append(items, &item)
	}

	if len(items) == 0 {
		return nil, errors.New("user-not-found")
	}

	return &GetUsers{
		Limit: int32(limit),
		List:  items,
	}, nil
}

// CreateUser for creating user
func (c *InitAPI) CreateUser(ctx context.Context, req *User, rolesID string) (*UserId, error) {
	var id string
	roles, err := c.GetRoles(rolesID)
	if err != nil {
		log.Println(err)
		if err.Error() == "no rows in result set" {
			return nil, errors.New("ERROR-NO-ADMIN-FOUND")
		}
		return nil, err
	}

	if roles != "ADMIN" {
		return nil, errors.New("invalid-roles")
	}

	// status := strconv.Itoa(req.Status)
	err = c.Db.QueryRow(`INSERT INTO users (username, email, status, role_id, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.Username, req.Email, req.Status, req.RoleId, time.Now()).Scan(&id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UserId{
		Id: id,
	}, nil
}

// GetRoles
func (c *InitAPI) GetRoles(id string) (string, error) {
	var roles string
	err := c.Db.QueryRow(`SELECT roles FROM roles WHERE id = $1`, id).Scan(&roles)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return roles, nil
}
