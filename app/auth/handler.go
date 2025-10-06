package auth

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		IsSuperuser *bool  `json:"is_superuser"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "could not hash password", http.StatusInternalServerError)
		return
	}

	isSuperuser := false
	claims, _ := GetClaimsFromContext(r.Context())
	if claims != nil && claims.IsSuperuser && req.IsSuperuser != nil {
		isSuperuser = *req.IsSuperuser
	}

	if err := CreateUser(req.Username, hashedPassword, isSuperuser); err != nil {
		http.Error(w, "user creation failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("user registered"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		ID:          user.ID,
		Username:    user.Username,
		IsSuperuser: user.IsSuperuser,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	tokenStr, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": tokenStr,
	})
}
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "missing username", http.StatusBadRequest)
		return
	}

	if err := DeleteUser(username); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("user deleted"))
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           claims.ID,
		"username":     claims.Username,
		"is_superuser": claims.IsSuperuser,
	})
}

func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := ListUsers()
	if err != nil {
		http.Error(w, "could not fetch users", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "missing username in path", http.StatusBadRequest)
		return
	}
	username := parts[3]

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	var req struct {
		NewUsername *string `json:"new_username"`
		Password    *string `json:"password"`
		IsSuperuser *bool   `json:"is_superuser"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.NewUsername != nil {
		user.Username = *req.NewUsername
	}
	if req.Password != nil {
		hashed, err := HashPassword(*req.Password)
		if err != nil {
			http.Error(w, "could not hash password", http.StatusInternalServerError)
			return
		}
		user.PasswordHash = hashed
	}
	if req.IsSuperuser != nil {
		user.IsSuperuser = *req.IsSuperuser
	}

	_, err = DB.Exec("UPDATE users SET username=?, password_hash=?, is_superuser=? WHERE id=?",
		user.Username, user.PasswordHash, user.IsSuperuser, user.ID)
	if err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("user updated"))
}
func UpdateOwnPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	usernameRaw := r.Context().Value("username")
	if usernameRaw == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	username := usernameRaw.(string)

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		http.Error(w, "invalid current password", http.StatusUnauthorized)
		return
	}

	hashed, err := HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "could not hash password", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec("UPDATE users SET password_hash=? WHERE username=?", hashed, username)
	if err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("password updated"))
}
