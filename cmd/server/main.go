package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"quan-ly-chi-tieu/internal/store"
)

type server struct{ store *store.Store }
type authedHandler func(http.ResponseWriter, *http.Request, store.PublicUser)

func main() {
	s, err := store.New(getenv("DATA_FILE", "data/expenses.json"))
	if err != nil {
		log.Fatalf("khởi tạo dữ liệu thất bại: %v", err)
	}
	app := &server{store: s}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/register", app.handleRegister)
	mux.HandleFunc("/api/auth/login", app.handleLogin)
	mux.HandleFunc("/api/auth/logout", app.auth(app.handleLogout))
	mux.HandleFunc("/api/me", app.auth(app.handleMe))
	mux.HandleFunc("/api/expenses", app.auth(app.handleExpenses))
	mux.HandleFunc("/api/expenses/", app.auth(app.handleExpenseByID))
	mux.HandleFunc("/api/summary", app.auth(app.handleSummary))
	mux.HandleFunc("/api/categories", app.auth(app.handleCategories))
	mux.HandleFunc("/api/categories/", app.auth(app.handleCategoryByID))
	mux.HandleFunc("/api/wallets", app.auth(app.handleWallets))
	mux.HandleFunc("/api/wallets/transfer", app.auth(app.handleTransfer))
	mux.HandleFunc("/api/budgets", app.auth(app.handleBudgets))
	mux.HandleFunc("/api/goals", app.auth(app.handleGoals))
	mux.HandleFunc("/api/debts", app.auth(app.handleDebts))
	mux.HandleFunc("/api/debts/", app.auth(app.handleDebtByID))
	mux.HandleFunc("/api/debts/complete/", app.auth(app.handleCompleteDebt))
	mux.HandleFunc("/api/import.csv", app.auth(app.handleImportCSV))
	mux.HandleFunc("/api/account", app.auth(app.handleAccount))
	mux.HandleFunc("/api/reminder", app.auth(app.handleReminder))
	mux.HandleFunc("/api/reminder/test", app.auth(app.handleReminderTest))
	mux.HandleFunc("/api/export.csv", app.auth(app.handleExportCSV))
	mux.HandleFunc("/api/export.xlsx", app.auth(app.handleExportExcel))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("/", spaHandler(getenv("STATIC_DIR", "frontend/dist")))
	addr := getenv("ADDR", ":8080")
	go app.runTelegramReminderLoop()
	log.Printf("Ứng dụng đang chạy tại http://localhost%s", addr)
	for _, url := range lanURLs(addr) {
		log.Printf("Thiết bị khác cùng Wi-Fi/LAN mở: %s", url)
	}
	log.Fatal(http.ListenAndServe(addr, logRequests(mux)))
}

func (s *server) auth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			if c, err := r.Cookie("token"); err == nil {
				token = c.Value
			}
		}
		user, ok := s.store.UserByToken(token)
		if !ok {
			writeError(w, http.StatusUnauthorized, "Vui lòng đăng nhập")
			return
		}
		next(w, r, user)
	}
}
func (s *server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "Phương thức không được hỗ trợ")
		return
	}
	var in store.RegisterInput
	if decode(w, r, &in) {
		u, t, err := s.store.Register(in)
		respondAuth(w, u, t, err)
	}
}
func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "Phương thức không được hỗ trợ")
		return
	}
	var in store.LoginInput
	if decode(w, r, &in) {
		u, t, err := s.store.Login(in)
		status := http.StatusBadRequest
		if errors.Is(err, store.ErrUnauthorized) {
			status = http.StatusUnauthorized
		}
		if err != nil {
			writeError(w, status, err.Error())
			return
		}
		respondAuth(w, u, t, nil)
	}
}
func (s *server) handleLogout(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	_ = s.store.Logout(token)
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
func (s *server) handleMe(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	writeJSON(w, 200, u)
}
func respondAuth(w http.ResponseWriter, u store.PublicUser, token string, err error) {
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"user": u, "token": token})
}

func (s *server) handleExpenses(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.List(u.ID, parseFilter(r)))
	case http.MethodPost:
		var in store.ExpenseInput
		if decode(w, r, &in) {
			e, err := s.store.Create(u.ID, in)
			if err != nil {
				writeError(w, 400, err.Error())
				return
			}
			writeJSON(w, 201, e)
		}
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) handleExpenseByID(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	id, ok := idFromPath(w, r, "/api/expenses/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPut:
		var in store.ExpenseInput
		if decode(w, r, &in) {
			e, err := s.store.Update(u.ID, id, in)
			writeResult(w, e, err)
		}
	case http.MethodDelete:
		err := s.store.Delete(u.ID, id)
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		w.WriteHeader(204)
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) handleSummary(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	writeJSON(w, 200, s.store.Summary(u.ID, parseFilter(r)))
}
func (s *server) handleCategories(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.Categories(u.ID))
	case http.MethodPost:
		var in store.CategoryInput
		if decode(w, r, &in) {
			c, err := s.store.UpsertCategory(u.ID, in)
			writeResult(w, c, err)
		}
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) handleCategoryByID(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	id, ok := idFromPath(w, r, "/api/categories/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPut:
		var in store.CategoryInput
		if decode(w, r, &in) {
			c, err := s.store.RenameCategory(u.ID, id, in)
			writeResult(w, c, err)
		}
	case http.MethodDelete:
		err := s.store.DeleteCategory(u.ID, id)
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		w.WriteHeader(204)
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) handleWallets(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.Wallets(u.ID))
	case http.MethodPost:
		var in store.WalletInput
		if decode(w, r, &in) {
			x, err := s.store.CreateWallet(u.ID, in)
			writeResult(w, x, err)
		}
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) handleTransfer(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	var in store.TransferInput
	if decode(w, r, &in) {
		err := s.store.Transfer(u.ID, in)
		if err != nil {
			writeError(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}
}
func (s *server) handleBudgets(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.Budgets(u.ID, r.URL.Query().Get("month")))
	case http.MethodPost:
		var in store.BudgetInput
		if decode(w, r, &in) {
			b, err := s.store.SaveBudget(u.ID, in)
			writeResult(w, b, err)
		}
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) handleGoals(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.Goals(u.ID))
	case http.MethodPost:
		var in store.GoalInput
		if decode(w, r, &in) {
			g, err := s.store.SaveGoal(u.ID, in)
			writeResult(w, g, err)
		}
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}

func (s *server) handleDebts(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.Debts(u.ID))
	case http.MethodPost:
		var in store.DebtInput
		if decode(w, r, &in) {
			d, err := s.store.SaveDebt(u.ID, in)
			writeResult(w, d, err)
		}
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) handleDebtByID(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	id, ok := idFromPath(w, r, "/api/debts/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodDelete:
		if err := s.store.DeleteDebt(u.ID, id); err != nil {
			writeStoreErr(w, err)
			return
		}
		writeJSON(w, 200, map[string]string{"status": "ok"})
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) handleCompleteDebt(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	id, ok := idFromPath(w, r, "/api/debts/complete/")
	if !ok {
		return
	}
	var body struct {
		WalletID int64 `json:"walletId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := s.store.CompleteDebt(u.ID, id, body.WalletID); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *server) handleReminderTest(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "Phương thức không được hỗ trợ")
		return
	}
	rem := s.store.Reminder(u.ID)
	if rem.TelegramChatID == "" {
		writeError(w, 400, "Vui lòng nhập Telegram Chat ID trước")
		return
	}
	msg := "⏰ Nhắc ghi chi tiêu: Hôm nay bạn có khoản thu/chi nào chưa ghi không? Mở app để cập nhật nhé."
	if err := sendTelegramMessage(rem.TelegramChatID, msg); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "sent"})
}

func (s *server) handleReminder(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.Reminder(u.ID))
	case http.MethodPost:
		var in store.ReminderInput
		if decode(w, r, &in) {
			rem, err := s.store.SaveReminder(u.ID, in)
			writeResult(w, rem, err)
		}
	default:
		writeError(w, 405, "Phương thức không được hỗ trợ")
	}
}
func (s *server) runTelegramReminderLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		s.sendDueTelegramReminders()
		<-ticker.C
	}
}
func (s *server) sendDueTelegramReminders() {
	token := getenv("TELEGRAM_BOT_TOKEN", "")
	if token == "" {
		return
	}
	now := time.Now()
	for _, rem := range s.store.DueReminders(now) {
		msg := "⏰ Nhắc ghi chi tiêu: Hôm nay bạn có khoản thu/chi nào chưa ghi không? Mở app để cập nhật nhé."
		if err := sendTelegramMessage(rem.TelegramChatID, msg); err == nil {
			_ = s.store.MarkReminderSent(rem.UserID, now.Format("2006-01-02"))
		}
	}
}
func sendTelegramMessage(chatID, msg string) error {
	token := getenv("TELEGRAM_BOT_TOKEN", "")
	if token == "" {
		return errors.New("Chưa cấu hình TELEGRAM_BOT_TOKEN")
	}
	form := url.Values{}
	form.Set("chat_id", chatID)
	form.Set("text", msg)
	resp, err := http.PostForm("https://api.telegram.org/bot"+token+"/sendMessage", form)
	if err != nil {
		return errors.New("Không gửi được Telegram: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return errors.New("Telegram từ chối gửi. Kiểm tra Chat ID hoặc token bot")
	}
	return nil
}

func urlQueryEscape(v string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(v, " ", "+"), ":", "%3A"), "/", "%2F")
}

func (s *server) handleAccount(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	if r.Method != http.MethodDelete {
		writeError(w, 405, "Phương thức không được hỗ trợ")
		return
	}
	if err := s.store.DeleteAccount(u.ID); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}
func (s *server) handleImportCSV(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "Phương thức không được hỗ trợ")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "Vui lòng chọn file CSV")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, 400, "Không đọc được file")
		return
	}
	records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		writeError(w, 400, "CSV không hợp lệ")
		return
	}
	count := 0
	for i, row := range records {
		if i == 0 && len(row) > 0 && strings.Contains(strings.ToLower(row[0]), "ng") {
			continue
		}
		if len(row) < 5 {
			continue
		}
		amount, _ := strconv.ParseFloat(strings.ReplaceAll(row[4], ",", ""), 64)
		if amount <= 0 {
			continue
		}
		in := store.ExpenseInput{Date: strings.TrimSpace(row[0]), Type: strings.TrimSpace(row[1]), Category: strings.TrimSpace(row[2]), Title: strings.TrimSpace(row[3]), Amount: amount, CostKind: "variable"}
		if len(row) > 5 {
			in.Note = strings.TrimSpace(row[5])
		}
		if len(row) > 6 {
			for _, tag := range strings.Split(row[6], ";") {
				in.Tags = append(in.Tags, strings.TrimSpace(tag))
			}
		}
		if len(row) > 7 {
			in.CostKind = strings.TrimSpace(row[7])
		}
		if _, err := s.store.Create(u.ID, in); err == nil {
			count++
		}
	}
	writeJSON(w, 200, map[string]int{"imported": count})
}

func (s *server) handleExportCSV(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=transactions.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Ngày", "Loại", "Danh mục", "Tên khoản", "Số tiền", "Ghi chú", "Tags", "Loại chi phí"})
	for _, e := range s.store.List(u.ID, parseFilter(r)) {
		_ = cw.Write([]string{e.Date, e.Type, e.Category, e.Title, fmt.Sprintf("%.0f", e.Amount), e.Note, strings.Join(e.Tags, ";"), e.CostKind})
	}
	cw.Flush()
}

func (s *server) handleExportExcel(w http.ResponseWriter, r *http.Request, u store.PublicUser) {
	file := excelize.NewFile()
	defer file.Close()

	sheet := "Giao dịch"
	file.SetSheetName("Sheet1", sheet)
	headers := []string{"Ngày", "Loại", "Danh mục", "Tên khoản", "Số tiền", "Ghi chú", "Tags", "Loại chi phí"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}

	moneyStyle, _ := file.NewStyle(&excelize.Style{NumFmt: 3})
	headerStyle, _ := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"2563EB"}, Pattern: 1},
	})
	_ = file.SetCellStyle(sheet, "A1", "H1", headerStyle)

	for row, e := range s.store.List(u.ID, parseFilter(r)) {
		r := row + 2
		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", r), e.Date)
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", r), e.Type)
		_ = file.SetCellValue(sheet, fmt.Sprintf("C%d", r), e.Category)
		_ = file.SetCellValue(sheet, fmt.Sprintf("D%d", r), e.Title)
		_ = file.SetCellValue(sheet, fmt.Sprintf("E%d", r), e.Amount)
		_ = file.SetCellValue(sheet, fmt.Sprintf("F%d", r), e.Note)
		_ = file.SetCellValue(sheet, fmt.Sprintf("G%d", r), strings.Join(e.Tags, ";"))
		_ = file.SetCellValue(sheet, fmt.Sprintf("H%d", r), e.CostKind)
	}
	_ = file.SetCellStyle(sheet, "E2", fmt.Sprintf("E%d", len(s.store.List(u.ID, parseFilter(r)))+1), moneyStyle)
	_ = file.SetColWidth(sheet, "A", "A", 14)
	_ = file.SetColWidth(sheet, "B", "C", 16)
	_ = file.SetColWidth(sheet, "D", "D", 28)
	_ = file.SetColWidth(sheet, "E", "E", 16)
	_ = file.SetColWidth(sheet, "F", "H", 24)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=transactions.xlsx")
	if err := file.Write(w); err != nil {
		writeError(w, 500, "Không xuất được file Excel")
	}
}
func parseFilter(r *http.Request) store.Filter {
	q := r.URL.Query()
	min, _ := strconv.ParseFloat(q.Get("minAmount"), 64)
	max, _ := strconv.ParseFloat(q.Get("maxAmount"), 64)
	wid, _ := strconv.ParseInt(q.Get("walletId"), 10, 64)
	return store.Filter{Category: strings.TrimSpace(q.Get("category")), Type: strings.TrimSpace(q.Get("type")), Search: strings.TrimSpace(q.Get("search")), Month: strings.TrimSpace(q.Get("month")), From: q.Get("from"), To: q.Get("to"), MinAmount: min, MaxAmount: max, WalletID: wid, Tag: strings.TrimSpace(q.Get("tag")), CostKind: strings.TrimSpace(q.Get("costKind"))}
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, 400, "Dữ liệu không hợp lệ")
		return false
	}
	return true
}
func idFromPath(w http.ResponseWriter, r *http.Request, prefix string) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, prefix), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, 400, "ID không hợp lệ")
		return 0, false
	}
	return id, true
}
func writeResult(w http.ResponseWriter, v any, err error) {
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, 200, v)
}
func writeStoreErr(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, 404, err.Error())
		return
	}
	writeError(w, 400, err.Error())
}
func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	b, _ := json.Marshal(v)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(status)
	_, _ = w.Write(b)
}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
func lanURLs(addr string) []string {
	port := "8080"
	if strings.HasPrefix(addr, ":") {
		port = strings.TrimPrefix(addr, ":")
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	urls := []string{}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}
			urls = append(urls, "http://"+ip.String()+":"+port)
		}
	}
	return urls
}
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
