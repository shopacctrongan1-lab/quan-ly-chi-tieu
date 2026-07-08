<template>
  <main :class="['shell', { dark }]">
    <button v-if="user" class="theme-toggle" type="button" @click="toggleTheme">{{ dark ? '☀️' : '🌙' }}</button>
    <div class="toast-stack" aria-live="polite">
      <div v-for="t in toasts" :key="t.id" class="toast" :class="t.type">
        <b>{{ t.title }}</b>
        <span>{{ t.message }}</span>
      </div>
    </div>

    <section v-if="!user" class="auth card">
      <h1>Quản lí chi tiêu</h1>
      <p class="muted">Đăng nhập để bảo mật dữ liệu tài chính cá nhân.</p>
      <form @submit.prevent="submitAuth">
        <input v-if="authMode === 'register'" v-model="auth.name" placeholder="Họ tên" required />
        <input v-model="auth.email" type="email" placeholder="Email" required />
        <input v-model="auth.password" type="password" placeholder="Mật khẩu" required />
        <button class="primary">{{ authMode === 'login' ? 'Đăng nhập' : 'Đăng ký' }}</button>
      </form>
      <button class="link" @click="authMode = authMode === 'login' ? 'register' : 'login'">
        {{ authMode === 'login' ? 'Chưa có tài khoản? Đăng ký' : 'Đã có tài khoản? Đăng nhập' }}
      </button>
      <p v-if="error" class="error">{{ error }}</p>
    </section>

    <template v-else>
      <section class="hero card">
        <div>
          <p class="eyebrow">Xin chào, {{ user.name }}</p>
          <h1>Dashboard tài chính</h1>
          <p class="muted">Quản lý thu chi, ví, nợ vay và mục tiêu trong từng khu vực riêng.</p>
        </div>
      </section>
      <button class="notif-bell" type="button" @click="toggleNotif" :aria-label="'Thông báo' + (unreadCount ? ' (' + unreadCount + ' chưa đọc)' : '')">
        🔔
        <span v-if="unreadCount > 0" class="notif-badge">{{ unreadCount }}</span>
      </button>
      <teleport to="body">
        <div v-if="notifOpen" class="notif-backdrop" @click="notifOpen = false"></div>
        <div v-if="notifOpen" class="notif-panel" role="dialog" aria-label="Danh sách thông báo">
          <header class="notif-head">Thông báo</header>
          <div class="notif-list">
            <button v-for="n in notifications" :key="n.id" class="notif-item" type="button" @click="gotoNotif(n)">
              <span class="notif-icon">{{ n.icon }}</span>
              <div class="notif-body">
                <b>{{ n.title }}</b>
                <p>{{ n.text }}</p>
              </div>
            </button>
            <div v-if="notifications.length === 0" class="notif-empty">Không có thông báo mới 🎉</div>
          </div>
        </div>
      </teleport>

      <button class="mobile-menu-toggle" type="button" @click="mobileMenuOpen = true">
        <span>☰</span>
        <b>{{ tabs.find(t => t.id === activeSection)?.label }}</b>
      </button>
      <div class="mobile-menu-backdrop" :class="{ open: mobileMenuOpen }" @click="mobileMenuOpen = false"></div>
      <nav class="section-tabs card" :class="{ open: mobileMenuOpen }" aria-label="Chọn mục quản lý">
        <div class="mobile-menu-head">
          <b>Menu quản lý</b>
          <button type="button" @click="mobileMenuOpen = false">×</button>
        </div>
        <button
          v-for="tab in tabs"
          :key="tab.id"
          :class="{ active: activeSection === tab.id }"
          type="button"
          @click="activeSection = tab.id; mobileMenuOpen = false"
        >
          <span class="tab-icon" aria-hidden="true">{{ tab.icon }}</span>
          <span class="tab-text">{{ tab.label }}</span>
        </button>
        <div class="mobile-menu-spacer"></div>
        <button class="mobile-sidebar-logout" type="button" @click="logout(); mobileMenuOpen = false">Đăng xuất</button>
      </nav>


      <section v-show="activeSection === 'dashboard'" class="stats-grid">
        <article class="stat card"><span>Số dư ví</span><strong>{{ displayMoney(walletBalance) }}</strong></article>
        <article class="stat card income"><span>Thu nhập tháng</span><strong>{{ displayMoney(summary.income) }}</strong></article>
        <article class="stat card expense"><span>Chi tiêu tháng</span><strong>{{ displayMoney(summary.expense) }}</strong></article>
        <article class="stat card balance"><span>Cân đối</span><strong>{{ displayMoney(summary.balance) }}</strong></article>
      </section>

      <section v-show="activeSection === 'dashboard'" class="insight-grid">
        <article class="insight-card card" v-for="item in dashboardInsights" :key="item.title" :class="item.tone">
          <span>{{ item.icon }}</span>
          <div>
            <b>{{ item.title }}</b>
            <p>{{ item.text }}</p>
          </div>
        </article>
      </section>
      <section v-show="activeSection === 'dashboard'" class="chart-controls card">
        <div>
          <h2>Khoảng thời gian báo cáo</h2>
          <p class="muted">Chọn ngày, tuần, tháng hoặc năm để 2 biểu đồ bên dưới tính đúng theo nhu cầu.</p>
        </div>
        <div class="chart-filter-row">
          <select v-model="chartPeriod" @change="saveChartPeriod">
            <option value="day">Theo ngày</option>
            <option value="week">Theo tuần</option>
            <option value="month">Theo tháng</option>
            <option value="year">Theo năm</option>
          </select>
          <label class="date-field"><span>{{ chartPeriod === 'day' ? 'Chọn ngày' : chartPeriod === 'week' ? 'Chọn ngày trong tuần' : chartPeriod === 'month' ? 'Chọn tháng' : 'Chọn năm' }}</span><input v-if="chartPeriod === 'month'" v-model="chartMonth" type="month" @change="syncChartDateFromMonth"/><input v-else-if="chartPeriod === 'year'" v-model="chartYear" type="number" min="2000" max="2100" @change="syncChartDateFromYear"/><input v-else v-model="chartDate" type="date"/></label>
          <button type="button" class="primary" @click="applyDashboardFilter">Lọc báo cáo</button>
        </div>
      </section>

      <section v-show="activeSection === 'dashboard'" class="charts-grid">
        <article class="card chart">
          <h2>Chi tiêu theo danh mục</h2>
          <template v-if="categoryRows.length">
            <div class="pie" :style="pieStyle"></div>
            <div class="legend"><span v-for="r in categoryRows" :key="r.name"><i :style="{background:r.color}"></i>{{ r.name }} - {{ displayMoney(r.value) }}</span></div>
          </template>
          <div v-else class="empty-chart">Chưa có khoản chi tiêu trong khoảng thời gian này. Hãy thêm giao dịch loại “Chi tiêu” để xem biểu đồ danh mục.</div>
        </article>
        <article class="card chart">
          <h2>So sánh thu / chi</h2>
          <template v-if="hasTrendData">
            <div class="bar-chart">
              <div class="bar-row income"><div class="bar-label"><b>Thu nhập</b><span>{{ displayMoney(summary.income) }}</span></div><div class="bar-track"><i :style="{width: incomeBarWidth}"></i></div></div>
              <div class="bar-row expense"><div class="bar-label"><b>Chi tiêu</b><span>{{ displayMoney(summary.expense) }}</span></div><div class="bar-track"><i :style="{width: expenseBarWidth}"></i></div></div>
              <div class="bar-result"><b>Còn lại</b><strong :class="summary.balance >= 0 ? 'income' : 'expense'">{{ displayMoney(summary.balance) }}</strong></div>
            </div>
            <p class="muted">Biểu đồ này cho biết trong khoảng thời gian đang chọn bạn thu bao nhiêu, chi bao nhiêu và còn lại bao nhiêu.</p>
          </template>
          <div v-else class="empty-chart">Chưa có giao dịch trong khoảng thời gian này. Hãy thêm thu nhập hoặc chi tiêu để xem so sánh.</div>
        </article>
      </section>

      <section v-show="activeSection === 'transactions'" class="transaction-workspace">
        <form class="card form transaction-form-only" @submit.prevent="submitExpense">
          <div class="section-head">
            <div>
              <p class="eyebrow">Nhập liệu</p>
              <h2 class="gradient-title">{{ editingId ? 'Sửa giao dịch' : 'Thêm giao dịch mới' }}</h2>
              <p class="muted">Ghi nhanh một khoản thu nhập hoặc chi tiêu.</p>
            </div>
            <span class="soft-badge">{{ form.type === 'income' ? 'Thu nhập' : 'Chi tiêu' }}</span>
          </div>
          <input v-model.trim="form.title" placeholder="Tên khoản giao dịch" required />
          <div class="two"><label class="money-field"><span>Số tiền giao dịch</span><input v-model="form.amount" type="text" inputmode="numeric" placeholder="Nhập 1 → chọn 10k / 100k / 1tr" required @blur="normalizeMoney(form, 'amount')"/><div class="money-presets"><button v-for="x in moneySuggestions(form.amount)" :key="x" type="button" @click="setMoney(form, 'amount', x)">{{ shortMoney(x) }}</button></div></label><label class="date-field"><span>Ngày giao dịch</span><input v-model="form.date" type="date" required /></label></div>
          <select v-model="form.type"><option value="expense">Chi tiêu / Tiền ra</option><option value="income">Thu nhập / Tiền vào</option></select>
          <input v-model.trim="form.category" placeholder="Nhập danh mục giao dịch" required />
          <textarea v-model.trim="form.note" placeholder="Ghi chú"></textarea>
          <div class="actions"><button class="primary">{{ editingId ? 'Cập nhật giao dịch' : 'Lưu giao dịch' }}</button><button v-if="editingId" type="button" @click="resetForm">Hủy sửa</button></div><p v-if="error" class="error">{{ error }}</p>
        </form>

        <section class="card quick-history">
          <div class="section-head">
            <div>
              <h2 class="gradient-title">Giao dịch gần đây</h2>
              <p class="muted">Xem nhanh các khoản vừa ghi. Muốn lọc theo ngày hoặc xuất Excel thì sang mục Truy vấn.</p>
            </div>
            <button type="button" class="query-shortcut" @click="activeSection='transactionQuery'">🔎 Truy vấn</button>
          </div>
          <div v-if="recentTransactions.length" class="quick-history-list">
            <article v-for="e in recentTransactions" :key="e.id" class="quick-history-row" :class="e.type">
              <div class="quick-history-icon">{{ e.type === 'income' ? '+' : '-' }}</div>
              <div class="quick-history-main">
                <div><b>{{ e.title }}</b><strong>{{ e.type==='income'?'+':'-' }}{{ displayMoney(e.amount) }}</strong></div>
                <small>{{ formatDate(e.date) }} • {{ e.category }} • {{ walletName(e.walletId) }}</small>
                <p v-if="e.note">{{ e.note }}</p>
              </div>
              <div class="quick-history-actions">
                <button type="button" @click="editItem(e)">Sửa</button>
                <button type="button" class="danger" @click="removeItem(e.id)">Xóa</button>
              </div>
            </article>
          </div>
          <div v-else class="empty-state compact quick-empty">
            <b>Chưa có giao dịch</b>
            <span>Sau khi lưu giao dịch, danh sách gần đây sẽ hiện ở đây.</span>
          </div>
        </section>
      </section>
      <section v-show="activeSection === 'transactionQuery'" class="bank-transactions">
        <section class="bank-query card">
          <div class="bank-query-head query-hero-head">
            <div class="query-title-block">
              <span class="query-kicker">Lịch sử tài khoản</span>
              <h2><span>Truy vấn</span> giao dịch</h2>
              <p>Chọn khoảng thời gian để xem tiền vào, tiền ra và rà soát các khoản đã ghi.</p>
            </div>
            <a class="button export-btn" :href="exportHref" target="_blank">📤 Xuất Excel</a>
          </div>
          <div class="bank-date-grid">
            <label class="bank-date-box">
              <span>Từ ngày</span>
              <input v-model="filters.from" type="date" />
            </label>
            <label class="bank-date-box">
              <span>Đến ngày</span>
              <input v-model="filters.to" type="date" />
            </label>
          </div>
          <button type="button" class="bank-query-btn" @click="loadData">Truy vấn</button>
          <p class="bank-note">Hệ thống hỗ trợ truy vấn lịch sử giao dịch đã lưu. Bạn có thể xuất kết quả ra Excel để dùng trên Microsoft Excel.</p>
          <div class="bank-extra-filters">
            <input v-model="filters.search" placeholder="Tìm theo tên hoặc ghi chú" @input="debouncedLoad" />
            <select v-model="filters.type" @change="loadData"><option value="">Tất cả thu/chi</option><option value="expense">Tiền ra</option><option value="income">Tiền vào</option></select>
            <select v-model="filters.costKind" @change="loadData"><option value="">Mọi chi phí</option><option value="fixed">Cố định</option><option value="variable">Biến đổi</option></select>
          </div>
        </section>

        <section class="bank-history">
          <div v-if="groupedTransactions.length" class="bank-day-list">
            <article v-for="group in groupedTransactions" :key="group.date" class="bank-day-card">
              <header>{{ formatDate(group.date) }}</header>
              <div v-for="e in group.items" :key="e.id" class="bank-row">
                <div class="bank-row-icon" :class="e.type">{{ e.type === 'income' ? '▣' : '↺' }}</div>
                <div class="bank-row-main">
                  <div class="bank-row-top">
                    <b>{{ e.type === 'income' ? 'TIỀN VÀO' : 'TIỀN RA' }}</b>
                    <strong :class="e.type">{{ e.type === 'income' ? '+' : '-' }}{{ plainMoney(e.amount) }} VND</strong>
                  </div>
                  <div class="bank-row-detail">
                    <span>{{ e.title || e.category }}</span>
                    <time>{{ formatTime(e.createdAt) }}</time>
                  </div>
                  <p>{{ e.note || e.category }} • {{ walletName(e.walletId) }} • {{ e.costKind === 'fixed' ? 'Cố định' : 'Biến đổi' }}</p>
                </div>
              </div>
            </article>
          </div>
          <div v-else class="empty-state card query-empty-state">
            <div class="empty-icon">🧾</div>
            <div>
              <b>Không có giao dịch trong khoảng thời gian này</b>
              <span>Hãy đổi ngày truy vấn hoặc thêm giao dịch mới.</span>
            </div>
          </div>
        </section>
      </section>
      <section v-show="activeSection === 'goals'" class="tools-grid goals-section">
        <article class="card goal-panel">
          <div class="section-head">
            <div>
              <h2 class="gradient-title">Tiết kiệm theo mục tiêu</h2>
              <p class="muted">Nhập thứ bạn muốn mua/làm, số tiền cần có và số tiền hiện có. Ứng dụng sẽ tính còn thiếu bao nhiêu và mỗi ngày nên để dành bao nhiêu.</p>
            </div>
          </div>
          <form class="goal-form" @submit.prevent="saveGoal">
            <input v-model="goal.name" placeholder="Tên mục tiêu, ví dụ: Mua laptop, đi du lịch, quỹ khẩn cấp"/>
            <label class="money-field"><span>Cần tổng cộng bao nhiêu tiền?</span><input v-model="goal.targetAmount" type="text" inputmode="numeric" placeholder="VD: nhập 20tr hoặc chọn gợi ý" @blur="normalizeMoney(goal, 'targetAmount')"/><div class="money-presets"><button v-for="x in moneySuggestions(goal.targetAmount)" :key="x" type="button" @click="setMoney(goal, 'targetAmount', x)">{{ shortMoney(x) }}</button></div></label>
            <label class="money-field"><span>Hiện tại đã có bao nhiêu?</span><input v-model="goal.currentAmount" type="text" inputmode="numeric" placeholder="VD: 5tr" @blur="normalizeMoney(goal, 'currentAmount')"/><div class="money-presets"><button v-for="x in moneySuggestions(goal.currentAmount)" :key="x" type="button" @click="setMoney(goal, 'currentAmount', x)">{{ shortMoney(x) }}</button></div></label>
            <label class="date-field"><span>Muốn hoàn thành trước ngày nào?</span><input v-model="goal.deadline" type="date"/></label>
            <button class="primary">{{ goal.name ? 'Lưu / cập nhật mục tiêu' : 'Tạo mục tiêu' }}</button>
          </form>
          <div v-if="goals.length" class="goal-list">
            <article v-for="g in goals" :key="g.id" class="goal-card" :class="goalRemaining(g) <= 0 ? 'done' : ''">
              <header>
                <div>
                  <b>{{ g.name }}</b>
                  <span>{{ goalStatusText(g) }}</span>
                </div>
                <strong>{{ Math.round(g.percent || 0) }}%</strong>
              </header>
              <div class="goal-track"><i :style="{width: Math.min(100, Math.max(0, g.percent || 0)) + '%'}"></i></div>
              <div class="goal-stats">
                <div><span>Đã có</span><b>{{ displayMoney(g.currentAmount) }}</b></div>
                <div><span>Còn thiếu</span><b>{{ displayMoney(goalRemaining(g)) }}</b></div>
                <div><span>Cần/ngày</span><b>{{ displayMoney(goalDailyNeed(g)) }}</b></div>
                <div><span>Cần/tháng</span><b>{{ displayMoney(g.monthlyNeed || 0) }}</b></div>
              </div>
              <div v-if="goalHistory(g).length" class="goal-history">
                <b>Lịch sử góp gần đây</b>
                <span v-for="h in goalHistory(g).slice(0,3)" :key="h.time">+{{ displayMoney(h.amount) }} • {{ formatDate(h.date) }}</span>
              </div>
              <footer>
                <button type="button" @click="addGoalMoney(g, 100000)">+100k</button>
                <button type="button" @click="addGoalMoney(g, 500000)">+500k</button>
                <button type="button" @click="addGoalMoney(g, 1000000)">+1tr</button>
                <button type="button" class="primary" @click="completeGoal(g)">Đã đạt</button>
                <button type="button" @click="editGoal(g)">Sửa</button>
                <button type="button" class="danger" @click="removeGoal(g)">Xóa</button>
              </footer>
            </article>
          </div>
          <div v-else class="empty-state compact goal-empty"><b>Chưa có mục tiêu tiết kiệm</b><span>Hãy tạo một mục tiêu để biết mỗi ngày cần để dành bao nhiêu.</span></div>
        </article>
      </section>
      <section v-show="activeSection === 'debts'" class="debt-workspace">
        <article class="card debt-form-card">
          <div class="section-head">
            <div>
              <p class="eyebrow">Ghi nhận</p>
              <h2 class="gradient-title">Nợ & cho vay</h2>
              <p class="muted">Theo dõi khoản bạn nợ ai hoặc ai đang nợ bạn.</p>
            </div>
            <span class="soft-badge">{{ debt.kind === 'borrow' ? 'Tôi nợ' : 'Cho vay' }}</span>
          </div>
          <form @submit.prevent="saveDebt">
            <select v-model="debt.kind">
              <option value="borrow">Tôi nợ</option>
              <option value="lend">Người khác nợ tôi</option>
            </select>
            <input v-model="debt.person" placeholder="Tên người liên quan" required />
            <label class="money-field">
              <span>Số tiền nợ / cho vay</span>
              <input v-model="debt.amount" type="text" inputmode="numeric" placeholder="Nhập 1 → chọn 10k / 100k / 1tr" @blur="normalizeMoney(debt, 'amount')" required />
              <div class="money-presets">
                <button v-for="x in moneySuggestions(debt.amount)" :key="x" type="button" @click="setMoney(debt, 'amount', x)">{{ shortMoney(x) }}</button>
              </div>
            </label>
            <label class="date-field">
              <span>Ngày hẹn trả / thu nợ</span>
              <input v-model="debt.dueDate" type="date" />
            </label>
            <textarea v-model="debt.note" placeholder="Ghi chú (tùy chọn)"></textarea>
            <button class="primary">Lưu khoản nợ</button>
          </form>
        </article>

        <article class="card debt-list-card">
          <h2 class="gradient-title">Danh sách nợ</h2>
          <div v-if="debts.length" class="debt-list">
            <div v-for="d in debts" :key="d.id" class="debt-row" :class="[d.kind, d.status === 'done' ? 'done' : '']">
              <div class="debt-row-icon">{{ d.kind === 'borrow' ? '↑' : '↓' }}</div>
              <div class="debt-row-main">
                <div class="debt-row-top">
                  <b>{{ d.kind === 'borrow' ? 'Tôi nợ' : 'Cho vay' }} — {{ d.person }}</b>
                  <strong :class="d.kind === 'borrow' ? 'expense' : 'income'">{{ displayMoney(d.amount) }}</strong>
                </div>
                <small>Hạn: {{ d.dueDate ? formatDate(d.dueDate) : 'không có' }} • <span :class="d.status === 'done' ? 'income' : 'muted'">{{ d.status === 'done' ? 'Đã hoàn thành' : 'Đang theo dõi' }}</span></small>
                <p v-if="d.note" class="muted">{{ d.note }}</p>
              </div>
              <div class="debt-actions">
                <button v-if="d.status !== 'done'" type="button" class="primary" @click="completeDebt(d)">{{ d.kind === 'borrow' ? 'Đã trả' : 'Đã thu' }}</button>
                <button type="button" class="danger" @click="removeDebt(d)">Xóa</button>
              </div>
            </div>
          </div>
          <div v-else class="empty-state compact">
            <b>Chưa có khoản nợ nào</b>
            <span>Lưu khoản nợ hoặc cho vay bên trái để theo dõi ở đây.</span>
          </div>
        </article>
      </section>
      <section v-show="activeSection === 'settings'" class="tools-grid"><article class="card"><h2>Nhắc nhở Telegram</h2><p class="muted">Chọn giờ nhắc mỗi ngày để bot Telegram nhắn bạn ghi lại chi tiêu.</p><form @submit.prevent="saveTelegramReminder"><label class="toggle-row"><input v-model="telegramReminder.enabled" type="checkbox" /><span>Bật nhắc nhở hằng ngày</span></label><label class="date-field"><span>Giờ nhắc mỗi ngày</span><input v-model="telegramReminder.time" type="time" /></label><input v-model.trim="telegramReminder.telegramChatId" placeholder="Nhập Chat ID, ví dụ: 5687993964"/><div class="actions"><button>Lưu nhắc nhở Telegram</button><button type="button" class="primary" @click="testTelegramReminder">Gửi thử bot</button></div></form><div class="guide-mini"><b>Cách lấy Chat ID:</b><ol><li>Tạo bot bằng @BotFather và cấu hình TELEGRAM_BOT_TOKEN trên server.</li><li>Nhắn /start cho bot của bạn.</li><li>Nhập Chat ID của bạn, không nhập @username của bot.</li><li>Bấm “Gửi thử bot” để kiểm tra ngay.</li></ol></div></article><article class="card"><h2>Cài đặt & riêng tư</h2><p class="muted">Import CSV mẫu: Ngày, Loại, Danh mục, Tên khoản, Số tiền, Ghi chú, Loại chi phí.</p><input type="file" accept=".csv" @change="importCsv"/><button class="danger" @click="deleteAccount">Xóa tài khoản và toàn bộ dữ liệu</button></article></section>
    </template>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { api, setToken } from './api'
const today = new Date().toISOString().slice(0,10), currentMonth = today.slice(0,7)
const tabs=[
  {id:'dashboard',icon:'📊',label:'Tổng quan'},
  {id:'transactions',icon:'💸',label:'Giao dịch'},
  {id:'transactionQuery',icon:'🔎',label:'Truy vấn'},
  {id:'goals',icon:'🐷',label:'Tiết kiệm'},
  {id:'debts',icon:'🤝',label:'Nợ vay'},
  {id:'settings',icon:'⚙️',label:'Cài đặt'}
]
const sectionHelp={
  dashboard:{title:'Tổng quan tài chính',desc:'Xem nhanh sức khỏe tài chính trong tháng đang chọn.',tips:['Số dư ví là tổng tiền hiện có trong tất cả ví.','Biểu đồ danh mục cho biết tiền chi nhiều nhất ở đâu.','Bấm “Ẩn số dư” khi dùng app ở nơi đông người.']},
  transactions:{title:'Giao dịch thu / chi',desc:'Chỉ dùng để thêm hoặc sửa khoản thu chi hằng ngày.',tips:['Nhập tên khoản, số tiền, ngày giao dịch và chọn đúng ví.','Chọn danh mục để báo cáo và ngân sách tính chính xác.','Nếu cần xem lịch sử hoặc xuất Excel, hãy sang mục Truy vấn.']},
  transactionQuery:{title:'Truy vấn giao dịch',desc:'Xem lại lịch sử giao dịch theo ngày và xuất Excel.',tips:['Chọn Từ ngày và Đến ngày rồi bấm Truy vấn.','Có thể lọc tiền vào/tiền ra hoặc tìm theo từ khóa.','Danh sách được gom theo từng ngày giống app ngân hàng.']},
  wallets:{title:'Ví và tài khoản',desc:'Quản lý tiền trong 2 nơi chính: Ví tiền mặt và Ngân hàng.',tips:['Ứng dụng chỉ dùng 2 ví cố định: Ví tiền mặt và Ngân hàng.','Mỗi giao dịch cần chọn tiền đi từ Ví tiền mặt hay Ngân hàng.','Bạn có thể tự đặt mức tiền tối thiểu để cảnh báo số dư thấp.']},
  goals:{title:'Mục tiêu tiết kiệm',desc:'Theo dõi mục tiêu lớn như mua laptop, du lịch, quỹ khẩn cấp.',tips:['Nhập số tiền mục tiêu, số đã có và hạn hoàn thành.','Hệ thống tính số tiền cần tiết kiệm mỗi tháng.']},
  debts:{title:'Nợ và cho vay',desc:'Theo dõi bạn nợ ai hoặc ai đang nợ bạn.',tips:['“Tôi nợ” sẽ trừ tiền ví khi bấm Đã trả nợ.','“Người khác nợ tôi” sẽ cộng tiền ví khi bấm Đã thu nợ.','Luôn chọn đúng ví dùng để trả hoặc nhận tiền.']},
  settings:{title:'Cài đặt & riêng tư',desc:'Nhập dữ liệu, bật nhắc nhở và quản lý quyền riêng tư.',tips:['Import CSV dùng để chuyển dữ liệu từ file cũ vào app.','Xóa tài khoản sẽ xóa toàn bộ dữ liệu và không thể hoàn tác.']}
}
const user=ref(null), authMode=ref('login'), activeSection=ref(localStorage.getItem('activeSection')||'dashboard'), chartPeriod=ref('month'), chartDate=ref(today), chartMonth=ref(currentMonth), chartYear=ref(new Date().getFullYear()), dark=ref(localStorage.getItem('dark')==='1'), hideMoney=ref(false), lowBalanceLimit=ref(Number(localStorage.getItem('lowBalanceLimit')||200000)), error=ref(''), editingId=ref(null), mobileMenuOpen=ref(false), toasts=ref([]), goalContribHistory=ref({}), dashboardFilterApplied=ref(false), notifOpen=ref(false), notifRead=ref([])
if(!tabs.some(t=>t.id===activeSection.value)) activeSection.value='dashboard'
watch(activeSection, v=>localStorage.setItem('activeSection', v))
const expenses=ref([]), wallets=ref([]), goals=ref([]), debts=ref([])
const summary=ref({income:0,expense:0,balance:0,byCategory:{},dailyIncome:{},dailyExpense:{}})
const auth=reactive({name:'',email:'',password:''}), filters=reactive({month:'',search:'',type:'',from:'',to:'',costKind:''})
const blank=()=>({title:'',amount:null,category:'',type:'expense',date:today,note:'',walletId:0,tags:[],costKind:'variable'}); const form=reactive(blank())
const newWallet=reactive({name:'',balance:null}), goal=reactive({name:'',targetAmount:null,currentAmount:null,deadline:''}), debt=reactive({kind:'borrow',person:'',amount:null,dueDate:'',note:'',walletId:0}), telegramReminder=reactive({enabled:false,time:'21:00',telegramChatId:''})
let timer=null
function ymd(d){return d.toISOString().slice(0,10)}
function dashboardQuery(){
  const q={...filters}
  delete q.from; delete q.to; delete q.month
  const d=new Date(chartDate.value+'T00:00:00')
  if(chartPeriod.value==='day'){
    q.from=chartDate.value; q.to=chartDate.value
  } else if(chartPeriod.value==='week'){
    const start=new Date(d); const day=(start.getDay()+6)%7; start.setDate(start.getDate()-day)
    const end=new Date(start); end.setDate(start.getDate()+6)
    q.from=ymd(start); q.to=ymd(end)
  } else if(chartPeriod.value==='month'){
    q.month=chartMonth.value || chartDate.value.slice(0,7)
  } else {
    const y=Number(chartYear.value)||new Date().getFullYear(); q.from=y+'-01-01'; q.to=y+'-12-31'
  }
  return q
}
function saveChartPeriod(){}
function syncChartDateFromMonth(){chartDate.value=(chartMonth.value||currentMonth)+'-01'}
function syncChartDateFromYear(){chartDate.value=(chartYear.value||new Date().getFullYear())+'-01-01'}
function applyDashboardFilter(){dashboardFilterApplied.value=true; loadData()}
const coreWallets=computed(()=>wallets.value.filter(w=>Number.isFinite(Number(w.balance))))
const storedWalletBalance=computed(()=>coreWallets.value.reduce((s,w)=>s+(Number(w.balance)||0),0))
const walletBalance=computed(()=>{
  const stored=storedWalletBalance.value
  const sm=summary.value||{}
  const hasTransactions=(Number(sm.income)||0)>0 || (Number(sm.expense)||0)>0
  return stored!==0 || !hasTransactions ? stored : (Number(sm.balance)||0)
}), lowWallets=computed(()=>coreWallets.value.filter(w=>(Number(w.balance)||0) < lowBalanceLimit.value)), exportHref=computed(()=>api.exportUrl(filters))
const colors=['#2563eb','#ef4444','#16a34a','#f59e0b','#8b5cf6','#06b6d4','#ec4899']
const categoryRows=computed(()=>Object.entries(summary.value.byCategory||{}).map(([name,value],i)=>({name,value,color:colors[i%colors.length]})))
const pieStyle=computed(()=>{let total=categoryRows.value.reduce((s,r)=>s+r.value,0), acc=0; if(!total)return {}; const parts=categoryRows.value.map(r=>{const a=acc/total*100; acc+=r.value; return `${r.color} ${a}% ${acc/total*100}%`}); return {background:`conic-gradient(${parts.join(',')})`}})
const recentTransactions=computed(()=>[...expenses.value].sort((a,b)=>String(b.date).localeCompare(String(a.date)) || new Date(b.createdAt||0)-new Date(a.createdAt||0)).slice(0,8))
const groupedTransactions=computed(()=>{
  const groups=[]
  const byDate=new Map()
  for(const e of expenses.value){
    const key=e.date || (e.createdAt||'').slice(0,10) || today
    if(!byDate.has(key)){byDate.set(key,{date:key,items:[]}); groups.push(byDate.get(key))}
    byDate.get(key).items.push(e)
  }
  return groups.sort((a,b)=>b.date.localeCompare(a.date))
})
const hasTrendData=computed(()=>summary.value.income>0 || summary.value.expense>0)
const maxCompare=computed(()=>Math.max(1, summary.value.income||0, summary.value.expense||0))
const incomeBarWidth=computed(()=>Math.max(3, ((summary.value.income||0)/maxCompare.value)*100)+'%')
const expenseBarWidth=computed(()=>Math.max(3, ((summary.value.expense||0)/maxCompare.value)*100)+'%')
const topCategory=computed(()=>categoryRows.value.length?[...categoryRows.value].sort((a,b)=>b.value-a.value)[0]:null)
const savingRate=computed(()=>summary.value.income>0?Math.round(((summary.value.income-summary.value.expense)/summary.value.income)*100):0)
const dashboardInsights=computed(()=>{
  const income=summary.value.income||0, expense=summary.value.expense||0, balance=summary.value.balance||0
  const items=[]
  if(income||expense){
    const spendRate=income>0?Math.round(expense/income*100):0
    items.push({icon:spendRate>80?'⚠️':'✅',title:'Tỷ lệ chi tiêu',text:income>0?`Bạn đã chi ${spendRate}% thu nhập trong kỳ này.`:'Kỳ này có chi tiêu nhưng chưa ghi nhận thu nhập.',tone:spendRate>80?'warn':'good'})
    items.push({icon:balance>=0?'💰':'🔻',title:'Cân đối kỳ này',text:balance>=0?`Bạn còn dư ${money(balance)}.`:`Bạn đang âm ${money(Math.abs(balance))}, nên giảm chi không cần thiết.`,tone:balance>=0?'good':'bad'})
  } else {
    items.push({icon:'📝',title:'Chưa có dữ liệu',text:'Hãy thêm giao dịch để hệ thống đưa ra nhận xét rõ hơn.',tone:'info'})
  }
  if(topCategory.value) items.push({icon:'📌',title:'Chi nhiều nhất',text:`Mục ${topCategory.value.name} đang chi nhiều nhất: ${money(topCategory.value.value)}.`,tone:'info'})
  items.push({icon:savingRate.value>=20?'🐷':'🎯',title:'Khả năng tiết kiệm',text:income>0?`Tỷ lệ để dành hiện khoảng ${savingRate.value}%.`:'Ghi thêm thu nhập để xem tỷ lệ tiết kiệm.',tone:savingRate.value>=20?'good':'info'})
  return items.slice(0,4)
})
const notifications=computed(()=>{
  const items=[]
  // Nợ/cho vay đến hạn hoặc quá hạn (trong vòng 3 ngày)
  for(const d of debts.value){
    if(d.status==='done'||!d.dueDate) continue
    const end=new Date(d.dueDate+'T23:59:59')
    const diffDays=Math.ceil((end-new Date())/86400000)
    if(diffDays<=3){
      const who=d.kind==='borrow'?'Bạn nợ':'Cho vay'
      const when=diffDays<0?`Quá hạn ${Math.abs(diffDays)} ngày`:diffDays===0?'Đến hạn hôm nay':`Còn ${diffDays} ngày`
      items.push({id:'debt-'+d.id,icon:d.kind==='borrow'?'🔴':'🟡',title:`${who} ${d.person} — ${displayMoney(d.amount)}`,text:when,section:'debts'})
    }
  }
  // Chưa ghi giao dịch hôm nay
  const hasToday=expenses.value.some(e=>e.date===today)
  if(!hasToday) items.push({id:'no-entry-'+today,icon:'📝',title:'Chưa ghi giao dịch hôm nay',text:'Hãy thêm ít nhất 1 khoản thu/chi để theo dõi tài chính.',section:'transactions'})
  return items
})
const unreadCount=computed(()=>notifications.value.filter(n=>!notifRead.value.includes(n.id)).length)
function notifKey(){return 'notifRead:'+(user.value?.id||user.value?.email||'guest')}
function loadNotifRead(){try{notifRead.value=JSON.parse(localStorage.getItem(notifKey())||'[]')}catch(_){notifRead.value=[]}}
function saveNotifRead(){localStorage.setItem(notifKey(),JSON.stringify(notifRead.value))}
function toggleNotif(){
  notifOpen.value=!notifOpen.value
  if(notifOpen.value){
    // đánh dấu tất cả hiện tại là đã đọc
    const ids=[...new Set([...notifRead.value,...notifications.value.map(n=>n.id)])]
    notifRead.value=ids
    saveNotifRead()
  }
}
function gotoNotif(item){activeSection.value=item.section; notifOpen.value=false}
async function submitAuth(){try{error.value=''; const res=authMode.value==='login'?await api.login(auth):await api.register(auth); setToken(res.token); user.value=res.user; loadGoalHistory(); loadNotifRead(); await loadData(); success(authMode.value==='login'?'Đăng nhập thành công':'Đăng ký thành công')}catch(e){fail(e,'Không đăng nhập được')}}
async function logout(){await api.logout().catch(()=>{}); setToken(''); user.value=null}
async function loadData(){try{const [ex,sm,ws,gs,ds,rem]=await Promise.all([api.expenses(filters),api.summary(dashboardFilterApplied.value ? dashboardQuery() : {}),api.wallets(),api.goals(),api.debts(),api.reminder()]); expenses.value=ex; summary.value=sm; wallets.value=ws; goals.value=gs; debts.value=ds; Object.assign(telegramReminder,{enabled:!!rem.enabled,time:rem.time||'21:00',telegramChatId:rem.telegramChatId||''}); if(!form.walletId&&coreWallets.value[0])form.walletId=coreWallets.value[0].id; if(!debt.walletId&&coreWallets.value[0])debt.walletId=coreWallets.value[0].id}catch(e){fail(e,'Không tải được dữ liệu')}}
function debouncedLoad(){clearTimeout(timer); timer=setTimeout(loadData,250)}
function toast(type,title,message){
  const id=Date.now()+Math.random()
  toasts.value.push({id,type,title,message})
  setTimeout(()=>{toasts.value=toasts.value.filter(t=>t.id!==id)},3600)
}
function success(message,title='Thành công'){toast('success',title,message)}
function fail(e,title='Có lỗi xảy ra'){const message=e?.message||String(e||'Vui lòng thử lại'); error.value=message; toast('error',title,message)}
function info(message,title='Thông báo'){toast('info',title,message)}
function goalHistoryKey(){return 'goalHistory:'+(user.value?.id||user.value?.email||'guest')}
function loadGoalHistory(){try{goalContribHistory.value=JSON.parse(localStorage.getItem(goalHistoryKey())||'{}')}catch(_){goalContribHistory.value={}}}
function saveGoalHistory(){localStorage.setItem(goalHistoryKey(),JSON.stringify(goalContribHistory.value))}
function goalHistory(g){return goalContribHistory.value[String(g.id||g.name)]||[]}
function rememberGoalContribution(g,amount){
  const key=String(g.id||g.name)
  const list=[{amount,date:today,time:new Date().toISOString()},...(goalContribHistory.value[key]||[])].slice(0,10)
  goalContribHistory.value={...goalContribHistory.value,[key]:list}
  saveGoalHistory()
}
function syncDarkClass(){document.documentElement.classList.toggle('dark-page', dark.value)}
function toggleTheme(){dark.value=!dark.value; localStorage.setItem('dark', dark.value?'1':'0'); syncDarkClass()}
function parseMoneyValue(v){
  if(v===null||v===undefined||v==='') return null
  if(typeof v==='number') return v
  let raw=String(v).toLowerCase().trim().replace(/\s/g,'').replace(/,/g,'.')
  let multiplier=1
  if(raw.endsWith('m')||raw.endsWith('tr')||raw.endsWith('triệu')||raw.endsWith('trieu')){multiplier=1000000; raw=raw.replace(/(m|tr|triệu|trieu)$/,'')}
  else if(raw.endsWith('k')||raw.endsWith('nghìn')||raw.endsWith('nghin')){multiplier=1000; raw=raw.replace(/(k|nghìn|nghin)$/,'')}
  const n=parseFloat(raw.replace(/[^0-9.]/g,''))
  return Number.isFinite(n)?Math.round(n*multiplier):null
}
function normalizeMoney(obj,key){obj[key]=parseMoneyValue(obj[key])}
function addMoney(obj,key,amount){obj[key]=(parseMoneyValue(obj[key])||0)+amount}
function setMoney(obj,key,amount){obj[key]=amount}
function setLowBalanceLimit(amount){lowBalanceLimit.value=amount; saveLowBalanceLimit()}
function moneyBase(v){
  if(v===null||v===undefined||v==='') return null
  const m=String(v).trim().match(/\d+(?:[.,]\d+)?/)
  return m?Number(m[0].replace(',','.')):null
}
function moneySuggestions(v){
  const n=moneyBase(v)
  if(!n) return [10000,100000,1000000]
  return [n*10000,n*100000,n*1000000].map(x=>Math.round(x))
}
function shortMoney(v){
  if(v>=1000000) return (v/1000000).toLocaleString('vi-VN')+'tr'
  if(v>=1000) return (v/1000).toLocaleString('vi-VN')+'k'
  return v.toLocaleString('vi-VN')+'đ'
}
function normalizeLimit(){lowBalanceLimit.value=parseMoneyValue(lowBalanceLimit.value)||0; saveLowBalanceLimit()}
function saveLowBalanceLimit(){localStorage.setItem('lowBalanceLimit', String(lowBalanceLimit.value||0))}
async function submitExpense(){try{normalizeMoney(form,'amount'); form.tags=[]; const editing=!!editingId.value; editing?await api.updateExpense(editingId.value,form):await api.createExpense(form); resetForm(); await loadData(); success(editing?'Đã cập nhật giao dịch':'Đã lưu giao dịch')}catch(e){fail(e,'Không lưu được giao dịch')}}
function editItem(e){Object.assign(form,{...e,tags:[]}); editingId.value=e.id; scrollTo({top:0,behavior:'smooth'})}
async function removeItem(id){if(confirm('Xóa giao dịch này?')){try{await api.deleteExpense(id); await loadData(); success('Đã xóa giao dịch')}catch(e){fail(e,'Không xóa được giao dịch')}}}
function resetForm(){Object.assign(form,blank()); form.walletId=coreWallets.value[0]?.id||0; editingId.value=null}
async function saveWallet(){await api.createWallet(newWallet); newWallet.name=''; newWallet.balance=null; await loadData()}
function goalRemaining(g){return Math.max(0,(Number(g.targetAmount)||0)-(Number(g.currentAmount)||0))}
function goalDaysLeft(g){
  if(!g.deadline) return 0
  const end=new Date(g.deadline+'T23:59:59')
  const diff=Math.ceil((end-new Date())/86400000)
  return Math.max(0,diff)
}
function goalDailyNeed(g){
  const remain=goalRemaining(g), days=goalDaysLeft(g)
  if(remain<=0) return 0
  return days>0 ? remain/days : remain
}
function goalStatusText(g){
  const remain=goalRemaining(g), days=goalDaysLeft(g)
  if(remain<=0) return 'Đã đạt mục tiêu 🎉'
  if(!g.deadline) return 'Còn thiếu '+money(remain)+' - chưa đặt hạn hoàn thành'
  if(days<=0) return 'Đã tới hạn, còn thiếu '+money(remain)
  return 'Còn '+days+' ngày, nên để dành '+money(goalDailyNeed(g))+'/ngày'
}
async function saveGoal(){try{normalizeMoney(goal,'targetAmount'); normalizeMoney(goal,'currentAmount'); await api.saveGoal(goal); Object.assign(goal,{name:'',targetAmount:null,currentAmount:null,deadline:''}); await loadData(); success('Đã lưu mục tiêu tiết kiệm')}catch(e){fail(e,'Không lưu được mục tiêu')}}
async function addGoalMoney(g,amount){try{await api.saveGoal({name:g.name,targetAmount:g.targetAmount,currentAmount:(Number(g.currentAmount)||0)+amount,deadline:g.deadline}); rememberGoalContribution(g,amount); await loadData(); success('Đã ghi nhận góp thêm '+money(amount))}catch(e){fail(e,'Không góp thêm được')}}
async function completeGoal(g){try{const add=Math.max(0,(Number(g.targetAmount)||0)-(Number(g.currentAmount)||0)); await api.saveGoal({name:g.name,targetAmount:g.targetAmount,currentAmount:g.targetAmount,deadline:g.deadline}); if(add>0) rememberGoalContribution(g,add); await loadData(); success('Chúc mừng, mục tiêu đã hoàn thành!')}catch(e){fail(e,'Không cập nhật được mục tiêu')}}
function editGoal(g){Object.assign(goal,{name:g.name,targetAmount:g.targetAmount,currentAmount:g.currentAmount,deadline:g.deadline||''}); scrollTo({top:0,behavior:'smooth'})}
async function removeGoal(g){if(confirm('Xóa mục tiêu "'+g.name+'"?')){try{await api.deleteGoal(g.id); await loadData(); success('Đã xóa mục tiêu tiết kiệm')}catch(e){fail(e,'Không xóa được mục tiêu')}}}
async function saveDebt(){try{normalizeMoney(debt,'amount'); debt.walletId=debt.walletId||coreWallets.value[0]?.id||0; await api.saveDebt(debt); Object.assign(debt,{kind:'borrow',person:'',amount:null,dueDate:'',note:'',walletId:coreWallets.value[0]?.id||0}); await loadData(); success('Đã lưu khoản nợ/cho vay')}catch(e){fail(e,'Không lưu được khoản nợ')}}
async function completeDebt(d){try{await api.completeDebt(d.id,d.walletId); await loadData(); success(d.kind==='borrow'?'Đã ghi nhận trả nợ':'Đã ghi nhận thu nợ')}catch(e){fail(e,'Không cập nhật được khoản nợ')}}
async function removeDebt(d){
  const note=d.status==='done'?'\nKhoản này đã hoàn thành nên khi xóa hệ thống sẽ hoàn tác lại số dư ví liên quan.':''
  if(confirm(`Xóa khoản nợ/cho vay của ${d.person}?${note}`)){try{await api.deleteDebt(d.id); await loadData(); success('Đã xóa khoản nợ/cho vay')}catch(e){fail(e,'Không xóa được khoản nợ')}}
}
async function importCsv(e){const file=e.target.files?.[0]; if(file){try{const r=await api.importCsv(file); await loadData(); success(`Đã import ${r.imported} giao dịch`)}catch(err){fail(err,'Không import được CSV')}}}
async function saveTelegramReminder(){try{await api.saveReminder(telegramReminder); success('Đã lưu nhắc nhở Telegram')}catch(e){fail(e,'Không lưu được nhắc nhở')}}
async function testTelegramReminder(){try{await api.saveReminder(telegramReminder); await api.testReminder(); success('Đã gửi tin nhắn test. Kiểm tra Telegram nhé!')}catch(e){fail(e,'Không gửi được tin test')}}
async function deleteAccount(){if(confirm('Xóa vĩnh viễn tài khoản và toàn bộ dữ liệu?')){await api.deleteAccount(); setToken(''); user.value=null}}
function displayMoney(v){return hideMoney.value?'******':money(v)}
function money(v){return new Intl.NumberFormat('vi-VN',{style:'currency',currency:'VND'}).format(v||0)}
function plainMoney(v){return new Intl.NumberFormat('vi-VN').format(v||0)}
function formatDate(v){return new Intl.DateTimeFormat('vi-VN').format(new Date(v))}
function formatTime(v){const d=v?new Date(v):new Date(); return d.toLocaleTimeString('vi-VN',{hour:'2-digit',minute:'2-digit'})}
function walletName(id){return wallets.value.find(w=>w.id===id)?.name || 'Ví'}
onMounted(async()=>{syncDarkClass(); try{const me=await api.me(); user.value=me; loadGoalHistory(); loadNotifRead(); await loadData()}catch(_){}})
</script>
