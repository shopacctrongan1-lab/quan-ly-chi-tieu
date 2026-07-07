package store

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("Không tìm thấy dữ liệu")
var ErrUnauthorized = errors.New("Thông tin đăng nhập không hợp lệ")

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt"`
}
type PublicUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
type ReminderSetting struct {
	UserID         int64  `json:"userId"`
	Enabled        bool   `json:"enabled"`
	Time           string `json:"time"`
	TelegramChatID string `json:"telegramChatId"`
	LastSentDate   string `json:"lastSentDate"`
}
type ReminderInput struct {
	Enabled        bool   `json:"enabled"`
	Time           string `json:"time"`
	TelegramChatID string `json:"telegramChatId"`
}
type Session struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}
type Category struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"userId"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}
type Wallet struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Budget struct {
	ID       int64   `json:"id"`
	UserID   int64   `json:"userId"`
	Category string  `json:"category"`
	Month    string  `json:"month"`
	Limit    float64 `json:"limit"`
	Spent    float64 `json:"spent"`
	Percent  float64 `json:"percent"`
	Status   string  `json:"status"`
}
type Debt struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	WalletID  int64     `json:"walletId"`
	Kind      string    `json:"kind"`
	Person    string    `json:"person"`
	Amount    float64   `json:"amount"`
	DueDate   string    `json:"dueDate"`
	Note      string    `json:"note"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type DebtInput struct {
	WalletID int64   `json:"walletId"`
	Kind     string  `json:"kind"`
	Person   string  `json:"person"`
	Amount   float64 `json:"amount"`
	DueDate  string  `json:"dueDate"`
	Note     string  `json:"note"`
}
type SavingsGoal struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"userId"`
	Name          string    `json:"name"`
	TargetAmount  float64   `json:"targetAmount"`
	CurrentAmount float64   `json:"currentAmount"`
	Deadline      string    `json:"deadline"`
	MonthlyNeed   float64   `json:"monthlyNeed"`
	Percent       float64   `json:"percent"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
type Expense struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	WalletID  int64     `json:"walletId"`
	Title     string    `json:"title"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	Type      string    `json:"type"`
	Date      string    `json:"date"`
	Note      string    `json:"note"`
	Tags      []string  `json:"tags"`
	CostKind  string    `json:"costKind"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type ExpenseInput struct {
	Title    string   `json:"title"`
	Amount   float64  `json:"amount"`
	Category string   `json:"category"`
	Type     string   `json:"type"`
	Date     string   `json:"date"`
	Note     string   `json:"note"`
	Tags     []string `json:"tags"`
	CostKind string   `json:"costKind"`
	WalletID int64    `json:"walletId"`
}
type Filter struct {
	Category  string
	Type      string
	Search    string
	Month     string
	From      string
	To        string
	MinAmount float64
	MaxAmount float64
	WalletID  int64
	Tag       string
	CostKind  string
}
type Summary struct {
	Income       float64            `json:"income"`
	Expense      float64            `json:"expense"`
	Balance      float64            `json:"balance"`
	Count        int                `json:"count"`
	ByCategory   map[string]float64 `json:"byCategory"`
	DailyIncome  map[string]float64 `json:"dailyIncome"`
	DailyExpense map[string]float64 `json:"dailyExpense"`
	DailyTotals  map[string]float64 `json:"dailyTotals"`
}
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type CategoryInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type WalletInput struct {
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}
type TransferInput struct {
	FromWalletID int64    `json:"fromWalletId"`
	ToWalletID   int64    `json:"toWalletId"`
	Amount       float64  `json:"amount"`
	Date         string   `json:"date"`
	Note         string   `json:"note"`
	Tags         []string `json:"tags"`
	CostKind     string   `json:"costKind"`
}
type BudgetInput struct {
	Category string  `json:"category"`
	Month    string  `json:"month"`
	Limit    float64 `json:"limit"`
}
type GoalInput struct {
	Name          string  `json:"name"`
	TargetAmount  float64 `json:"targetAmount"`
	CurrentAmount float64 `json:"currentAmount"`
	Deadline      string  `json:"deadline"`
}

type Store struct{ db *sql.DB }

func New(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_journal_mode=DELETE&_busy_timeout=3000&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := s.seedIfEmpty(); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS wallets (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		balance REAL NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		wallet_id INTEGER NOT NULL DEFAULT 0,
		title TEXT NOT NULL,
		amount REAL NOT NULL,
		category TEXT NOT NULL,
		type TEXT NOT NULL,
		date TEXT NOT NULL,
		note TEXT NOT NULL DEFAULT '',
		tags TEXT NOT NULL DEFAULT '',
		cost_kind TEXT NOT NULL DEFAULT 'variable',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS budgets (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		category TEXT NOT NULL,
		month TEXT NOT NULL,
		lim REAL NOT NULL,
		UNIQUE(user_id, category, month)
	);
	CREATE TABLE IF NOT EXISTS debts (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		wallet_id INTEGER NOT NULL DEFAULT 0,
		kind TEXT NOT NULL,
		person TEXT NOT NULL,
		amount REAL NOT NULL,
		due_date TEXT NOT NULL DEFAULT '',
		note TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'open',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS goals (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		target_amount REAL NOT NULL,
		current_amount REAL NOT NULL DEFAULT 0,
		deadline TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE IF NOT EXISTS reminders (
		user_id INTEGER PRIMARY KEY,
		enabled INTEGER NOT NULL DEFAULT 0,
		time TEXT NOT NULL DEFAULT '21:00',
		telegram_chat_id TEXT NOT NULL DEFAULT '',
		last_sent_date TEXT NOT NULL DEFAULT ''
	);
	`)
	return err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(password))
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(h.Sum(nil)), nil
}

func checkPassword(stored, password string) bool {
	parts := strings.SplitN(stored, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(password))
	return subtle.ConstantTimeCompare(h.Sum(nil), expected) == 1
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func tagsToString(tags []string) string {
	var cleaned []string
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			cleaned = append(cleaned, t)
		}
	}
	return strings.Join(cleaned, ";")
}

func stringToTags(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ";")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}

func cleanTags(tags []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}

func cleanCostKind(s string) string {
	s = strings.TrimSpace(s)
	if s == "fixed" || s == "variable" || s == "investment" {
		return s
	}
	return "variable"
}

func validateExpense(in ExpenseInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return errors.New("Tên khoản không được để trống")
	}
	if in.Amount <= 0 {
		return errors.New("Số tiền phải lớn hơn 0")
	}
	t := strings.TrimSpace(in.Type)
	if t != "income" && t != "expense" {
		return errors.New("Loại phải là income hoặc expense")
	}
	return nil
}

func signedAmount(e Expense) float64 {
	if e.Type == "income" {
		return e.Amount
	}
	return -e.Amount
}

func matches(e Expense, f Filter) bool {
	if f.Category != "" && !strings.EqualFold(e.Category, f.Category) {
		return false
	}
	if f.Type != "" && !strings.EqualFold(e.Type, f.Type) {
		return false
	}
	if f.CostKind != "" && !strings.EqualFold(e.CostKind, f.CostKind) {
		return false
	}
	if f.WalletID != 0 && e.WalletID != f.WalletID {
		return false
	}
	if f.Search != "" {
		q := strings.ToLower(f.Search)
		if !strings.Contains(strings.ToLower(e.Title), q) &&
			!strings.Contains(strings.ToLower(e.Category), q) &&
			!strings.Contains(strings.ToLower(e.Note), q) {
			return false
		}
	}
	if f.Month != "" && !strings.HasPrefix(e.Date, f.Month) {
		return false
	}
	if f.From != "" && e.Date < f.From {
		return false
	}
	if f.To != "" && e.Date > f.To {
		return false
	}
	if f.MinAmount > 0 && e.Amount < f.MinAmount {
		return false
	}
	if f.MaxAmount > 0 && e.Amount > f.MaxAmount {
		return false
	}
	if f.Tag != "" {
		found := false
		for _, tg := range e.Tags {
			if strings.EqualFold(tg, f.Tag) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func enrichGoal(g SavingsGoal) SavingsGoal {
	if g.TargetAmount > 0 {
		g.Percent = g.CurrentAmount / g.TargetAmount * 100
		if g.Percent > 100 {
			g.Percent = 100
		}
	}
	if g.Deadline != "" {
		t, err := time.Parse("2006-01-02", g.Deadline)
		if err == nil {
			remaining := g.TargetAmount - g.CurrentAmount
			if remaining < 0 {
				remaining = 0
			}
			days := time.Until(t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)).Hours() / 24
			if days < 1 {
				days = 1
			}
			if remaining > 0 {
				months := days / 30
				if months < 1 {
					months = 1
				}
				g.MonthlyNeed = remaining / months
			}
		}
	}
	return g
}

// ── seed ─────────────────────────────────────────────────────────────────────

func (s *Store) seedIfEmpty() error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := hashPassword("demo123456")
	if err != nil {
		return err
	}
	now := time.Now()
	res, err := s.db.Exec(
		"INSERT INTO users(name,email,password_hash,created_at) VALUES(?,?,?,?)",
		"Demo User", "demo@example.com", hash, now,
	)
	if err != nil {
		return err
	}
	uid, _ := res.LastInsertId()

	cats := []struct{ name, t string }{
		{"Ăn uống", "expense"}, {"Di chuyển", "expense"}, {"Mua sắm", "expense"},
		{"Giải trí", "expense"}, {"Y tế", "expense"}, {"Giáo dục", "expense"},
		{"Hóa đơn", "expense"}, {"Lương", "income"}, {"Thưởng", "income"},
		{"Đầu tư", "income"},
	}
	for _, c := range cats {
		_, _ = s.db.Exec("INSERT INTO categories(user_id,name,type) VALUES(?,?,?)", uid, c.name, c.t)
	}

	wRes, err := s.db.Exec(
		"INSERT INTO wallets(user_id,name,balance,created_at,updated_at) VALUES(?,?,?,?,?)",
		uid, "Tiền mặt", 5000000, now, now,
	)
	if err != nil {
		return err
	}
	wid, _ := wRes.LastInsertId()

	_, _ = s.db.Exec("INSERT INTO wallets(user_id,name,balance,created_at,updated_at) VALUES(?,?,?,?,?)",
		uid, "Ngân hàng", 20000000, now, now)

	samples := []struct {
		title, cat, typ, date string
		amount               float64
	}{
		{"Cơm trưa", "Ăn uống", "expense", now.Format("2006-01-02"), 45000},
		{"Lương tháng", "Lương", "income", now.Format("2006-01-02"), 15000000},
		{"Xăng xe", "Di chuyển", "expense", now.AddDate(0, 0, -1).Format("2006-01-02"), 80000},
	}
	for _, e := range samples {
		_, _ = s.db.Exec(
			`INSERT INTO expenses(user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at)
			 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
			uid, wid, e.title, e.amount, e.cat, e.typ, e.date, "", "", "variable", now, now,
		)
		if e.typ == "income" {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance+?, updated_at=? WHERE id=?", e.amount, now, wid)
		} else {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance-?, updated_at=? WHERE id=?", e.amount, now, wid)
		}
	}
	return nil
}

// ── auth ──────────────────────────────────────────────────────────────────────

func (s *Store) Register(in RegisterInput) (PublicUser, string, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.Password = strings.TrimSpace(in.Password)
	if in.Name == "" {
		return PublicUser{}, "", errors.New("Tên không được để trống")
	}
	if in.Email == "" {
		return PublicUser{}, "", errors.New("Email không được để trống")
	}
	if len(in.Password) < 6 {
		return PublicUser{}, "", errors.New("Mật khẩu phải có ít nhất 6 ký tự")
	}
	hash, err := hashPassword(in.Password)
	if err != nil {
		return PublicUser{}, "", err
	}
	now := time.Now()
	res, err := s.db.Exec(
		"INSERT INTO users(name,email,password_hash,created_at) VALUES(?,?,?,?)",
		in.Name, in.Email, hash, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return PublicUser{}, "", errors.New("Email đã được sử dụng")
		}
		return PublicUser{}, "", err
	}
	uid, _ := res.LastInsertId()
	// seed default categories for new user
	cats := []struct{ name, t string }{
		{"Ăn uống", "expense"}, {"Di chuyển", "expense"}, {"Mua sắm", "expense"},
		{"Giải trí", "expense"}, {"Y tế", "expense"}, {"Hóa đơn", "expense"},
		{"Lương", "income"}, {"Thưởng", "income"},
	}
	for _, c := range cats {
		_, _ = s.db.Exec("INSERT INTO categories(user_id,name,type) VALUES(?,?,?)", uid, c.name, c.t)
	}
	_, _ = s.db.Exec("INSERT INTO wallets(user_id,name,balance,created_at,updated_at) VALUES(?,?,?,?,?)",
		uid, "Tiền mặt", 0, now, now)
	token := randomToken()
	_, _ = s.db.Exec("INSERT INTO sessions(token,user_id,created_at) VALUES(?,?,?)", token, uid, now)
	return PublicUser{ID: uid, Name: in.Name, Email: in.Email}, token, nil
}

func (s *Store) Login(in LoginInput) (PublicUser, string, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	var u User
	err := s.db.QueryRow(
		"SELECT id,name,email,password_hash,created_at FROM users WHERE email=?", in.Email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return PublicUser{}, "", ErrUnauthorized
	}
	if err != nil {
		return PublicUser{}, "", err
	}
	if !checkPassword(u.PasswordHash, in.Password) {
		return PublicUser{}, "", ErrUnauthorized
	}
	token := randomToken()
	_, _ = s.db.Exec("INSERT INTO sessions(token,user_id,created_at) VALUES(?,?,?)", token, u.ID, time.Now())
	return PublicUser{ID: u.ID, Name: u.Name, Email: u.Email}, token, nil
}

func (s *Store) Logout(token string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE token=?", token)
	return err
}

func (s *Store) UserByToken(token string) (PublicUser, bool) {
	if token == "" {
		return PublicUser{}, false
	}
	var uid int64
	if err := s.db.QueryRow("SELECT user_id FROM sessions WHERE token=?", token).Scan(&uid); err != nil {
		return PublicUser{}, false
	}
	var u PublicUser
	if err := s.db.QueryRow("SELECT id,name,email FROM users WHERE id=?", uid).Scan(&u.ID, &u.Name, &u.Email); err != nil {
		return PublicUser{}, false
	}
	return u, true
}

// ── expenses ──────────────────────────────────────────────────────────────────

func scanExpense(row *sql.Row) (Expense, error) {
	var e Expense
	var tagsStr string
	err := row.Scan(&e.ID, &e.UserID, &e.WalletID, &e.Title, &e.Amount, &e.Category,
		&e.Type, &e.Date, &e.Note, &tagsStr, &e.CostKind, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return e, err
	}
	e.Tags = stringToTags(tagsStr)
	return e, nil
}

func scanExpenseRows(rows *sql.Rows) (Expense, error) {
	var e Expense
	var tagsStr string
	err := rows.Scan(&e.ID, &e.UserID, &e.WalletID, &e.Title, &e.Amount, &e.Category,
		&e.Type, &e.Date, &e.Note, &tagsStr, &e.CostKind, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return e, err
	}
	e.Tags = stringToTags(tagsStr)
	return e, nil
}

func (s *Store) List(userID int64, f Filter) []Expense {
	rows, err := s.db.Query(
		`SELECT id,user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at
		 FROM expenses WHERE user_id=? ORDER BY date DESC, id DESC`, userID)
	if err != nil {
		return []Expense{}
	}
	defer rows.Close()
	var out []Expense
	for rows.Next() {
		e, err := scanExpenseRows(rows)
		if err != nil {
			continue
		}
		if matches(e, f) {
			out = append(out, e)
		}
	}
	if out == nil {
		return []Expense{}
	}
	return out
}

func (s *Store) Create(userID int64, in ExpenseInput) (Expense, error) {
	if err := validateExpense(in); err != nil {
		return Expense{}, err
	}
	in.Tags = cleanTags(in.Tags)
	in.CostKind = cleanCostKind(in.CostKind)
	now := time.Now()
	res, err := s.db.Exec(
		`INSERT INTO expenses(user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, in.WalletID, strings.TrimSpace(in.Title), in.Amount, strings.TrimSpace(in.Category),
		strings.TrimSpace(in.Type), strings.TrimSpace(in.Date), in.Note,
		tagsToString(in.Tags), in.CostKind, now, now,
	)
	if err != nil {
		return Expense{}, err
	}
	id, _ := res.LastInsertId()
	// update wallet balance
	if in.WalletID != 0 {
		if strings.TrimSpace(in.Type) == "income" {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance+?,updated_at=? WHERE id=? AND user_id=?",
				in.Amount, now, in.WalletID, userID)
		} else {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance-?,updated_at=? WHERE id=? AND user_id=?",
				in.Amount, now, in.WalletID, userID)
		}
	}
	e, err := scanExpense(s.db.QueryRow(
		`SELECT id,user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at
		 FROM expenses WHERE id=?`, id))
	return e, err
}

func (s *Store) Update(userID, id int64, in ExpenseInput) (Expense, error) {
	if err := validateExpense(in); err != nil {
		return Expense{}, err
	}
	// fetch old for wallet reversal
	old, err := scanExpense(s.db.QueryRow(
		`SELECT id,user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at
		 FROM expenses WHERE id=? AND user_id=?`, id, userID))
	if err == sql.ErrNoRows {
		return Expense{}, ErrNotFound
	}
	if err != nil {
		return Expense{}, err
	}
	in.Tags = cleanTags(in.Tags)
	in.CostKind = cleanCostKind(in.CostKind)
	now := time.Now()
	_, err = s.db.Exec(
		`UPDATE expenses SET wallet_id=?,title=?,amount=?,category=?,type=?,date=?,note=?,tags=?,cost_kind=?,updated_at=?
		 WHERE id=? AND user_id=?`,
		in.WalletID, strings.TrimSpace(in.Title), in.Amount, strings.TrimSpace(in.Category),
		strings.TrimSpace(in.Type), strings.TrimSpace(in.Date), in.Note,
		tagsToString(in.Tags), in.CostKind, now, id, userID,
	)
	if err != nil {
		return Expense{}, err
	}
	// reverse old wallet effect
	if old.WalletID != 0 {
		if old.Type == "income" {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance-?,updated_at=? WHERE id=? AND user_id=?",
				old.Amount, now, old.WalletID, userID)
		} else {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance+?,updated_at=? WHERE id=? AND user_id=?",
				old.Amount, now, old.WalletID, userID)
		}
	}
	// apply new wallet effect
	if in.WalletID != 0 {
		if strings.TrimSpace(in.Type) == "income" {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance+?,updated_at=? WHERE id=? AND user_id=?",
				in.Amount, now, in.WalletID, userID)
		} else {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance-?,updated_at=? WHERE id=? AND user_id=?",
				in.Amount, now, in.WalletID, userID)
		}
	}
	return scanExpense(s.db.QueryRow(
		`SELECT id,user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at
		 FROM expenses WHERE id=?`, id))
}

func (s *Store) Delete(userID, id int64) error {
	old, err := scanExpense(s.db.QueryRow(
		`SELECT id,user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at
		 FROM expenses WHERE id=? AND user_id=?`, id, userID))
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if _, err := s.db.Exec("DELETE FROM expenses WHERE id=? AND user_id=?", id, userID); err != nil {
		return err
	}
	now := time.Now()
	if old.WalletID != 0 {
		if old.Type == "income" {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance-?,updated_at=? WHERE id=? AND user_id=?",
				old.Amount, now, old.WalletID, userID)
		} else {
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance+?,updated_at=? WHERE id=? AND user_id=?",
				old.Amount, now, old.WalletID, userID)
		}
	}
	return nil
}

func (s *Store) Summary(userID int64, f Filter) Summary {
	expenses := s.List(userID, f)
	sum := Summary{
		ByCategory:   map[string]float64{},
		DailyIncome:  map[string]float64{},
		DailyExpense: map[string]float64{},
		DailyTotals:  map[string]float64{},
	}
	for _, e := range expenses {
		sum.Count++
		if e.Type == "income" {
			sum.Income += e.Amount
			sum.DailyIncome[e.Date] += e.Amount
		} else {
			sum.Expense += e.Amount
			sum.DailyExpense[e.Date] += e.Amount
			sum.ByCategory[e.Category] += e.Amount
		}
		sum.DailyTotals[e.Date] += signedAmount(e)
	}
	sum.Balance = sum.Income - sum.Expense
	return sum
}

// ── categories ────────────────────────────────────────────────────────────────

func (s *Store) Categories(userID int64) []Category {
	rows, err := s.db.Query("SELECT id,user_id,name,type FROM categories WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		return []Category{}
	}
	defer rows.Close()
	var out []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type); err == nil {
			out = append(out, c)
		}
	}
	if out == nil {
		return []Category{}
	}
	return out
}

func (s *Store) UpsertCategory(userID int64, in CategoryInput) (Category, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Type = strings.TrimSpace(in.Type)
	if in.Name == "" {
		return Category{}, errors.New("Tên danh mục không được để trống")
	}
	if in.Type != "income" && in.Type != "expense" {
		return Category{}, errors.New("Loại phải là income hoặc expense")
	}
	var existing Category
	err := s.db.QueryRow("SELECT id,user_id,name,type FROM categories WHERE user_id=? AND name=?",
		userID, in.Name).Scan(&existing.ID, &existing.UserID, &existing.Name, &existing.Type)
	if err == nil {
		return existing, nil
	}
	res, err := s.db.Exec("INSERT INTO categories(user_id,name,type) VALUES(?,?,?)", userID, in.Name, in.Type)
	if err != nil {
		return Category{}, err
	}
	id, _ := res.LastInsertId()
	return Category{ID: id, UserID: userID, Name: in.Name, Type: in.Type}, nil
}

func (s *Store) RenameCategory(userID, id int64, in CategoryInput) (Category, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return Category{}, errors.New("Tên danh mục không được để trống")
	}
	var old Category
	err := s.db.QueryRow("SELECT id,user_id,name,type FROM categories WHERE id=? AND user_id=?",
		id, userID).Scan(&old.ID, &old.UserID, &old.Name, &old.Type)
	if err == sql.ErrNoRows {
		return Category{}, ErrNotFound
	}
	if err != nil {
		return Category{}, err
	}
	newType := old.Type
	if in.Type == "income" || in.Type == "expense" {
		newType = in.Type
	}
	if _, err := s.db.Exec("UPDATE categories SET name=?,type=? WHERE id=? AND user_id=?",
		in.Name, newType, id, userID); err != nil {
		return Category{}, err
	}
	// update expenses with old category name
	if old.Name != in.Name {
		_, _ = s.db.Exec("UPDATE expenses SET category=?,updated_at=? WHERE user_id=? AND category=?",
			in.Name, time.Now(), userID, old.Name)
	}
	return Category{ID: id, UserID: userID, Name: in.Name, Type: newType}, nil
}

func (s *Store) DeleteCategory(userID, id int64) error {
	res, err := s.db.Exec("DELETE FROM categories WHERE id=? AND user_id=?", id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ── wallets ───────────────────────────────────────────────────────────────────

func scanWallet(row *sql.Row) (Wallet, error) {
	var w Wallet
	err := row.Scan(&w.ID, &w.UserID, &w.Name, &w.Balance, &w.CreatedAt, &w.UpdatedAt)
	return w, err
}

func (s *Store) Wallets(userID int64) []Wallet {
	rows, err := s.db.Query(
		"SELECT id,user_id,name,balance,created_at,updated_at FROM wallets WHERE user_id=? ORDER BY created_at", userID)
	if err != nil {
		return []Wallet{}
	}
	defer rows.Close()
	var out []Wallet
	for rows.Next() {
		var w Wallet
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Balance, &w.CreatedAt, &w.UpdatedAt); err == nil {
			out = append(out, w)
		}
	}
	if out == nil {
		return []Wallet{}
	}
	return out
}

func (s *Store) CreateWallet(userID int64, in WalletInput) (Wallet, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return Wallet{}, errors.New("Tên ví không được để trống")
	}
	now := time.Now()
	res, err := s.db.Exec(
		"INSERT INTO wallets(user_id,name,balance,created_at,updated_at) VALUES(?,?,?,?,?)",
		userID, in.Name, in.Balance, now, now,
	)
	if err != nil {
		return Wallet{}, err
	}
	id, _ := res.LastInsertId()
	return scanWallet(s.db.QueryRow(
		"SELECT id,user_id,name,balance,created_at,updated_at FROM wallets WHERE id=?", id))
}

func (s *Store) Transfer(userID int64, in TransferInput) error {
	if in.Amount <= 0 {
		return errors.New("Số tiền phải lớn hơn 0")
	}
	if in.FromWalletID == in.ToWalletID {
		return errors.New("Ví nguồn và đích không được trùng nhau")
	}
	now := time.Now()
	date := in.Date
	if date == "" {
		date = now.Format("2006-01-02")
	}
	// debit from source
	_, err := s.db.Exec("UPDATE wallets SET balance=balance-?,updated_at=? WHERE id=? AND user_id=?",
		in.Amount, now, in.FromWalletID, userID)
	if err != nil {
		return err
	}
	// credit to dest
	_, err = s.db.Exec("UPDATE wallets SET balance=balance+?,updated_at=? WHERE id=? AND user_id=?",
		in.Amount, now, in.ToWalletID, userID)
	if err != nil {
		return err
	}
	// record as transfer expense/income pair
	tags := tagsToString(cleanTags(in.Tags))
	ck := cleanCostKind(in.CostKind)
	note := in.Note
	_, _ = s.db.Exec(
		`INSERT INTO expenses(user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, in.FromWalletID, "Chuyển tiền", in.Amount, "Chuyển khoản", "expense", date, note, tags, ck, now, now,
	)
	_, _ = s.db.Exec(
		`INSERT INTO expenses(user_id,wallet_id,title,amount,category,type,date,note,tags,cost_kind,created_at,updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		userID, in.ToWalletID, "Nhận chuyển tiền", in.Amount, "Chuyển khoản", "income", date, note, tags, ck, now, now,
	)
	return nil
}

// ── budgets ───────────────────────────────────────────────────────────────────

func (s *Store) Budgets(userID int64, month string) []Budget {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	rows, err := s.db.Query(
		"SELECT id,user_id,category,month,lim FROM budgets WHERE user_id=? AND month=?", userID, month)
	if err != nil {
		return []Budget{}
	}
	defer rows.Close()
	var out []Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.Category, &b.Month, &b.Limit); err != nil {
			continue
		}
		var spent float64
		_ = s.db.QueryRow(
			"SELECT COALESCE(SUM(amount),0) FROM expenses WHERE user_id=? AND category=? AND type='expense' AND substr(date,1,7)=?",
			userID, b.Category, month,
		).Scan(&spent)
		b.Spent = spent
		if b.Limit > 0 {
			b.Percent = spent / b.Limit * 100
		}
		switch {
		case b.Percent >= 100:
			b.Status = "over"
		case b.Percent >= 80:
			b.Status = "warning"
		default:
			b.Status = "ok"
		}
		out = append(out, b)
	}
	if out == nil {
		return []Budget{}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Category < out[j].Category })
	return out
}

func (s *Store) SaveBudget(userID int64, in BudgetInput) (Budget, error) {
	in.Category = strings.TrimSpace(in.Category)
	if in.Category == "" {
		return Budget{}, errors.New("Danh mục không được để trống")
	}
	if in.Month == "" {
		in.Month = time.Now().Format("2006-01")
	}
	if in.Limit <= 0 {
		return Budget{}, errors.New("Hạn mức phải lớn hơn 0")
	}
	var id int64
	err := s.db.QueryRow(
		"SELECT id FROM budgets WHERE user_id=? AND category=? AND month=?",
		userID, in.Category, in.Month,
	).Scan(&id)
	if err == sql.ErrNoRows {
		res, err2 := s.db.Exec(
			"INSERT INTO budgets(user_id,category,month,lim) VALUES(?,?,?,?)",
			userID, in.Category, in.Month, in.Limit,
		)
		if err2 != nil {
			return Budget{}, err2
		}
		id, _ = res.LastInsertId()
	} else if err != nil {
		return Budget{}, err
	} else {
		if _, err := s.db.Exec("UPDATE budgets SET lim=? WHERE id=?", in.Limit, id); err != nil {
			return Budget{}, err
		}
	}
	list := s.Budgets(userID, in.Month)
	for _, b := range list {
		if b.ID == id {
			return b, nil
		}
	}
	return Budget{}, ErrNotFound
}

// ── goals ─────────────────────────────────────────────────────────────────────

func (s *Store) Goals(userID int64) []SavingsGoal {
	rows, err := s.db.Query(
		`SELECT id,user_id,name,target_amount,current_amount,deadline,created_at,updated_at
		 FROM goals WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return []SavingsGoal{}
	}
	defer rows.Close()
	var out []SavingsGoal
	for rows.Next() {
		var g SavingsGoal
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount,
			&g.Deadline, &g.CreatedAt, &g.UpdatedAt); err == nil {
			out = append(out, enrichGoal(g))
		}
	}
	if out == nil {
		return []SavingsGoal{}
	}
	return out
}

func (s *Store) SaveGoal(userID int64, in GoalInput) (SavingsGoal, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return SavingsGoal{}, errors.New("Tên mục tiêu không được để trống")
	}
	if in.TargetAmount <= 0 {
		return SavingsGoal{}, errors.New("Số tiền mục tiêu phải lớn hơn 0")
	}
	now := time.Now()
	var id int64
	err := s.db.QueryRow("SELECT id FROM goals WHERE user_id=? AND name=?", userID, in.Name).Scan(&id)
	if err == sql.ErrNoRows {
		res, err2 := s.db.Exec(
			`INSERT INTO goals(user_id,name,target_amount,current_amount,deadline,created_at,updated_at)
			 VALUES(?,?,?,?,?,?,?)`,
			userID, in.Name, in.TargetAmount, in.CurrentAmount, in.Deadline, now, now,
		)
		if err2 != nil {
			return SavingsGoal{}, err2
		}
		id, _ = res.LastInsertId()
	} else if err != nil {
		return SavingsGoal{}, err
	} else {
		if _, err := s.db.Exec(
			"UPDATE goals SET target_amount=?,current_amount=?,deadline=?,updated_at=? WHERE id=? AND user_id=?",
			in.TargetAmount, in.CurrentAmount, in.Deadline, now, id, userID,
		); err != nil {
			return SavingsGoal{}, err
		}
	}
	var g SavingsGoal
	err = s.db.QueryRow(
		`SELECT id,user_id,name,target_amount,current_amount,deadline,created_at,updated_at FROM goals WHERE id=?`, id,
	).Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Deadline, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return SavingsGoal{}, err
	}
	return enrichGoal(g), nil
}

// ── debts ─────────────────────────────────────────────────────────────────────

func (s *Store) Debts(userID int64) []Debt {
	rows, err := s.db.Query(
		`SELECT id,user_id,wallet_id,kind,person,amount,due_date,note,status,created_at,updated_at
		 FROM debts WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return []Debt{}
	}
	defer rows.Close()
	var out []Debt
	for rows.Next() {
		var d Debt
		if err := rows.Scan(&d.ID, &d.UserID, &d.WalletID, &d.Kind, &d.Person,
			&d.Amount, &d.DueDate, &d.Note, &d.Status, &d.CreatedAt, &d.UpdatedAt); err == nil {
			out = append(out, d)
		}
	}
	if out == nil {
		return []Debt{}
	}
	return out
}

func (s *Store) SaveDebt(userID int64, in DebtInput) (Debt, error) {
	in.Person = strings.TrimSpace(in.Person)
	if in.Person == "" {
		return Debt{}, errors.New("Tên người không được để trống")
	}
	if in.Amount <= 0 {
		return Debt{}, errors.New("Số tiền phải lớn hơn 0")
	}
	if in.Kind != "borrow" && in.Kind != "lend" {
		return Debt{}, errors.New("Loại phải là borrow hoặc lend")
	}
	now := time.Now()
	res, err := s.db.Exec(
		`INSERT INTO debts(user_id,wallet_id,kind,person,amount,due_date,note,status,created_at,updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		userID, in.WalletID, in.Kind, in.Person, in.Amount, in.DueDate, in.Note, "open", now, now,
	)
	if err != nil {
		return Debt{}, err
	}
	id, _ := res.LastInsertId()
	var d Debt
	err = s.db.QueryRow(
		`SELECT id,user_id,wallet_id,kind,person,amount,due_date,note,status,created_at,updated_at
		 FROM debts WHERE id=?`, id,
	).Scan(&d.ID, &d.UserID, &d.WalletID, &d.Kind, &d.Person,
		&d.Amount, &d.DueDate, &d.Note, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func (s *Store) DeleteDebt(userID, id int64) error {
	var d Debt
	err := s.db.QueryRow(
		`SELECT id,user_id,wallet_id,kind,person,amount,due_date,note,status,created_at,updated_at
		 FROM debts WHERE id=? AND user_id=?`, id, userID,
	).Scan(&d.ID, &d.UserID, &d.WalletID, &d.Kind, &d.Person,
		&d.Amount, &d.DueDate, &d.Note, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if _, err := s.db.Exec("DELETE FROM debts WHERE id=? AND user_id=?", id, userID); err != nil {
		return err
	}
	// reverse wallet adjustment if debt was completed
	if d.Status == "done" && d.WalletID != 0 {
		now := time.Now()
		if d.Kind == "borrow" {
			// borrow done: wallet was debited; restore it
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance+?,updated_at=? WHERE id=? AND user_id=?",
				d.Amount, now, d.WalletID, userID)
		} else {
			// lend done: wallet was credited; reverse it
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance-?,updated_at=? WHERE id=? AND user_id=?",
				d.Amount, now, d.WalletID, userID)
		}
	}
	return nil
}

func (s *Store) CompleteDebt(userID, id, walletID int64) error {
	var d Debt
	err := s.db.QueryRow(
		`SELECT id,user_id,wallet_id,kind,person,amount,due_date,note,status,created_at,updated_at
		 FROM debts WHERE id=? AND user_id=?`, id, userID,
	).Scan(&d.ID, &d.UserID, &d.WalletID, &d.Kind, &d.Person,
		&d.Amount, &d.DueDate, &d.Note, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if d.Status == "done" {
		return errors.New("Khoản nợ đã được hoàn thành")
	}
	now := time.Now()
	wid := walletID
	if wid == 0 {
		wid = d.WalletID
	}
	if _, err := s.db.Exec("UPDATE debts SET status='done',wallet_id=?,updated_at=? WHERE id=? AND user_id=?",
		wid, now, id, userID); err != nil {
		return err
	}
	if wid != 0 {
		if d.Kind == "borrow" {
			// paying back borrowed money — subtract from wallet
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance-?,updated_at=? WHERE id=? AND user_id=?",
				d.Amount, now, wid, userID)
		} else {
			// received repayment of lent money — add to wallet
			_, _ = s.db.Exec("UPDATE wallets SET balance=balance+?,updated_at=? WHERE id=? AND user_id=?",
				d.Amount, now, wid, userID)
		}
	}
	return nil
}

// ── reminders ─────────────────────────────────────────────────────────────────

func (s *Store) Reminder(userID int64) ReminderSetting {
	var r ReminderSetting
	var enabled int
	err := s.db.QueryRow(
		"SELECT user_id,enabled,time,telegram_chat_id,last_sent_date FROM reminders WHERE user_id=?", userID,
	).Scan(&r.UserID, &enabled, &r.Time, &r.TelegramChatID, &r.LastSentDate)
	if err != nil {
		return ReminderSetting{UserID: userID, Time: "21:00"}
	}
	r.Enabled = enabled == 1
	return r
}

func (s *Store) SaveReminder(userID int64, in ReminderInput) (ReminderSetting, error) {
	enabled := 0
	if in.Enabled {
		enabled = 1
	}
	t := strings.TrimSpace(in.Time)
	if t == "" {
		t = "21:00"
	}
	_, err := s.db.Exec(
		`INSERT INTO reminders(user_id,enabled,time,telegram_chat_id,last_sent_date)
		 VALUES(?,?,?,?,?)
		 ON CONFLICT(user_id) DO UPDATE SET enabled=excluded.enabled,time=excluded.time,telegram_chat_id=excluded.telegram_chat_id`,
		userID, enabled, t, strings.TrimSpace(in.TelegramChatID), "",
	)
	if err != nil {
		return ReminderSetting{}, err
	}
	return s.Reminder(userID), nil
}

func (s *Store) DueReminders(now time.Time) []ReminderSetting {
	hhmm := now.Format("15:04")
	today := now.Format("2006-01-02")
	rows, err := s.db.Query(
		`SELECT user_id,enabled,time,telegram_chat_id,last_sent_date FROM reminders
		 WHERE enabled=1 AND telegram_chat_id!='' AND time=? AND last_sent_date!=?`,
		hhmm, today,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []ReminderSetting
	for rows.Next() {
		var r ReminderSetting
		var enabled int
		if err := rows.Scan(&r.UserID, &enabled, &r.Time, &r.TelegramChatID, &r.LastSentDate); err == nil {
			r.Enabled = enabled == 1
			out = append(out, r)
		}
	}
	return out
}

func (s *Store) MarkReminderSent(userID int64, date string) error {
	_, err := s.db.Exec("UPDATE reminders SET last_sent_date=? WHERE user_id=?", date, userID)
	return err
}

// ── account ───────────────────────────────────────────────────────────────────

func (s *Store) DeleteAccount(userID int64) error {
	tables := []string{"expenses", "categories", "wallets", "budgets", "debts", "goals", "reminders", "sessions"}
	for _, tbl := range tables {
		if _, err := s.db.Exec("DELETE FROM "+tbl+" WHERE user_id=?", userID); err != nil {
			return err
		}
	}
	_, err := s.db.Exec("DELETE FROM users WHERE id=?", userID)
	return err
}
