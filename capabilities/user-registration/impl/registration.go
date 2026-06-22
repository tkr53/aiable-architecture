// Package registration は user-registration capability の実装である。
//
// この impl は spec/acceptance.md を満たすために生成された使い捨ての成果物である。
// 真の source は spec と test であり、この実装はいつでも再生成されうる。
// impl を直接編集して spec と乖離させてはならない。
package registration

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// 登録失敗の理由。spec の EMAIL_TAKEN / VALIDATION_ERROR に対応する。
var (
	ErrEmailTaken       = errors.New("EMAIL_TAKEN")
	ErrValidationFailed = errors.New("VALIDATION_ERROR")
)

// Request は登録要求。
type Request struct {
	Email    string
	Password string
}

// record は保存されるユーザーレコード。平文 password は保持しない。
type record struct {
	email        string // 入力表記をそのまま保存する
	passwordHash string // salt + hash。平文は復元できない
}

// Store はユーザーの保存先。正規化 email をキーにする。
type Store struct {
	users map[string]record
}

// NewStore は空のストアを返す。
func NewStore() *Store {
	return &Store{users: make(map[string]record)}
}

// Count は登録済みユーザー数を返す。
func (s *Store) Count() int {
	return len(s.users)
}

// normalize は email の同一性判定に使う正規化を行う（小文字化・前後空白除去）。
func normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Register は登録を試みる。成功すると nil を返す。
//
// AC-001: 正規化後に同一の email が既にあれば ErrEmailTaken。件数は増えない。
// AC-002: 有効な新規 email は成功し、件数が 1 増える。
// AC-003: 失敗時はストアを一切変更しない（早期 return で保証）。
// AC-005: password は salt 付き hash で保存し、平文は保持しない。
func Register(s *Store, r Request) error {
	key := normalize(r.Email)

	// 入力検証。失敗時はストアを変更しない。
	if key == "" || !strings.Contains(key, "@") || r.Password == "" {
		return ErrValidationFailed
	}

	// 重複判定は正規化後に行う。
	if _, exists := s.users[key]; exists {
		return ErrEmailTaken
	}

	hash, err := hashPassword(r.Password)
	if err != nil {
		return ErrValidationFailed
	}

	s.users[key] = record{
		email:        r.Email,
		passwordHash: hash,
	}
	return nil
}

// hashPassword は salt を付けて password をハッシュする。
// 同じ password でも salt が異なるため、毎回異なるハッシュになる（AC-005）。
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(password))
	digest := h.Sum(nil)
	// 保存形式: salt(hex) + ":" + digest(hex)。平文 password は含まれない。
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(digest), nil
}

// --- 以下はテストがストアの内部状態を検査するための、最小限の公開アクセサ ---

// NormalizedEmails は登録済みの正規化 email 集合を返す（AC-004 の状態比較用）。
func (s *Store) NormalizedEmails() map[string]struct{} {
	out := make(map[string]struct{}, len(s.users))
	for k := range s.users {
		out[k] = struct{}{}
	}
	return out
}

// StoredHash は指定した正規化 email の保存ハッシュを返す（AC-005 の検査用）。
// 存在しなければ第 2 戻り値が false。
func (s *Store) StoredHash(normalizedEmail string) (string, bool) {
	rec, ok := s.users[normalizedEmail]
	return rec.passwordHash, ok
}
