package authhandlers

import (
 "encoding/json"
 "net/http"

 "url-shortener/internal/auth"
)

type LoginRequest struct {
 User     string json:"user"
 Password string json:"password"
}

type LoginResponse struct {
 Token string json:"token"
}

type LoginHandler struct {
 tm       *auth.TokenManager
 adminUser string
 adminPass string
}

func NewLoginHandler(tm *auth.TokenManager, user, pass string) *LoginHandler {
 return &LoginHandler{tm: tm, adminUser: user, adminPass: pass}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 var req LoginRequest
 if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
  http.Error(w, "bad json", http.StatusBadRequest)
  return
 }

 if req.User != h.adminUser || req.Password != h.adminPass {
  http.Error(w, "invalid credentials", http.StatusUnauthorized)
  return
 }

 t, err := h.tm.NewToken(req.User)
 if err != nil {
  http.Error(w, "cannot issue token", http.StatusInternalServerError)
  return
 }

 w.Header().Set("Content-Type", "application/json")
 _ = json.NewEncoder(w).Encode(LoginResponse{Token: t})
}