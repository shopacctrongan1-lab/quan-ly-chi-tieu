<template>
  <main :class="['shell', { dark }]">
    <button v-if="user" class="theme-toggle" type="button" @click="toggleTheme">{{ dark ? '☀️' : '🌙' }}</button>

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
      <p class="hint">Demo: demo@example.com / demo123456</p>
      <p v-if="error" class="error">{{ error }}</p>
    </section>

    <template v-else>
      <section class="hero card">
        <div>
          <p class="eyebrow">Xin chào, {{ user.name }}</p>
          <h1>Dashboard tài chính</h1>
          <p class="muted">Quản lý thu chi, ví, nợ vay, ngân sách và mục tiêu trong từng khu vực riêng.</p>
        </div>
        <div class="top-actions">
          <button class="hero-logout" @click="logout">Đăng xuất</button>
        </div>
      </section>

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

      <section v-show="activeSection === 'dashboard'" class="chart-controls card">
        <div>
          <h2>Khoảng thời gian báo cáo</h2>
          <p class="muted">Chọn ngày, tuần, tháng hoặc năm để 2 biểu đồ bên dưới tính đúng theo nhu cầu.</p>
        </div>
        <div class="chart-filter-row">
          <select v-model="chartPeriod" @change="saveChartPeriod(); loadData()">
            <option value="day">Theo ngày</option>
            <option value="week">Theo tuần</option>
            <option value="month">Theo tháng</option>
            <option value="year">Theo năm</option>
          </select>
          <label class="date-field"><span>{{ chartPeriod === 'day' ? 'Chọn ngày' : chartPeriod === 'week' ? 'Chọn ngày trong tuần' : chartPeriod === 'month' ? 'Chọn tháng' : 'Chọn năm' }}</span><input v-if="chartPeriod === 'month'" v-model="chartMonth" type="month" @change="syncChartDateFromMonth"/><input v-else-if="chartPeriod === 'year'" v-model="chartYear" type="number" min="2000" max="2100" @change="syncChartDateFromYear"/><input v-else v-model="chartDate" type="date" @change="loadData"/></label>
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
              <h2>{{ editingId ? 'Sửa giao dịch' : 'Thêm giao dịch mới' }}</h2>
              <p class="muted">Ghi nhanh một khoản thu nhập hoặc chi tiêu.</p>
            </div>
            <span class="soft-badge">{{ form.type === 'income' ? 'Thu nhập' : 'Chi tiêu' }}</span>
          </div>
          <input v-model.trim="form.title" placeholder="Tên khoản giao dịch" required />
          <div class="two"><label class="money-field"><span>Số tiền giao dịch</span><input v-model="form.amount" type="text" inputmode="numeric" placeholder="Nhập 1 → chọn 10k / 100k / 1tr" required @blur="normalizeMoney(form, 'amount')"/><div class="money-presets"><button v-for="x in moneySuggestions(form.amount)" :key="x" type="button" @click="setMoney(form, 'amount', x)">{{ shortMoney(x) }}</button></div></label><label class="date-field"><span>Ngày giao dịch</span><input v-model="form.date" type="date" required /></label></div>
          <div class="two"><select v-model="form.type"><option value="expense">Chi tiêu / Tiền ra</option><option value="income">Thu nhập / Tiền vào</option></select><select v-model="form.walletId"><option v-for="w in coreWallets" :key="w.id" :value="w.id">{{ w.name }}</option></select></div>
          <input v-model.trim="form.category" list="categories" placeholder="Chọn danh mục giao dịch" required />
          <datalist id="categories"><option v-for="c in categories" :key="c.id" :value="c.name" /></datalist>
          <select v-model="form.costKind"><option value="variable">Chi phí biến đổi</option><option value="fixed">Chi phí cố định</option></select>
          <textarea v-model.trim="form.note" placeholder="Ghi chú"></textarea>
          <div class="actions"><button class="primary">{{ editingId ? 'Cập nhật giao dịch' : 'Lưu giao dịch' }}</button><button v-if="editingId" type="button" @click="resetForm">Hủy sửa</button></div><p v-if="error" class="error">{{ error }}</p>
        </form>

        <section class="card quick-history">
          <div class="section-head">
            <div>
              <h2>Giao dịch gần đây</h2>
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
                  <div class="bank-row-actions">
                    <button type="button" @click="editItem(e)">Sửa</button>
                    <button type="button" class="danger" @click="removeItem(e.id)">Xóa</button>
                  </div>
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
      <section v-show="activeSection === 'categories'" class="tools-grid"><article class="card"><h2>Danh mục</h2><form @submit.prevent="saveCategory"><input v-model="newCategory.name" placeholder="Tên danh mục"/><select v-model="newCategory.type"><option value="expense">Chi</option><option value="income">Thu</option></select><button>Thêm</button></form><p v-for="c in categories" :key="c.id" class="chip"><span>{{ c.name }}</span><span class="chip-actions"><button type="button" @click="renameCategory(c)">Đổi tên</button><button type="button" class="danger" @click="removeCategory(c)">Xóa</button></span></p></article></section>
      <section v-show="activeSection === 'wallets'" class="tools-grid"><article class="card"><h2>Ví</h2><p class="muted">Bạn chỉ dùng 2 ví cố định bên dưới. Khi thêm giao dịch, hãy chọn tiền đi từ Ví tiền mặt hay Ngân hàng.</p><div class="limit-box"><label><span>Mức cảnh báo số dư thấp</span><input v-model="lowBalanceLimit" type="text" inputmode="numeric" placeholder="Nhập 2 → chọn 20k / 200k / 2tr" @blur="normalizeLimit" /><div class="money-presets"><button v-for="x in moneySuggestions(lowBalanceLimit)" :key="x" type="button" @click="setLowBalanceLimit(x)">{{ shortMoney(x) }}</button></div></label><small>Ví nào có số dư thấp hơn mức này sẽ hiện cảnh báo.</small></div><div v-if="lowWallets.length" class="warning">Cảnh báo số dư thấp: {{ lowWallets.map(w => w.name).join(', ') }}</div><div class="wallet-grid"><div v-for="w in coreWallets" :key="w.id" class="wallet-card"><span>{{ w.name === 'Ví tiền mặt' ? '💵' : '🏦' }}</span><div><b>{{ w.name }}</b><strong>{{ displayMoney(w.balance) }}</strong><small>Cảnh báo khi dưới {{ displayMoney(lowBalanceLimit) }}</small></div></div></div></article></section>
      <section v-show="activeSection === 'budgets'" class="tools-grid"><article class="card"><h2>Ngân sách</h2><form @submit.prevent="saveBudget"><input v-model.trim="budget.category" list="categories" placeholder="Chọn danh mục cần đặt ngân sách" required /><label class="money-field"><span>Hạn mức ngân sách</span><input v-model="budget.limit" type="text" inputmode="numeric" placeholder="Nhập 2 → chọn 20k / 200k / 2tr" @blur="normalizeMoney(budget, 'limit')"/><div class="money-presets"><button v-for="x in moneySuggestions(budget.limit)" :key="x" type="button" @click="setMoney(budget, 'limit', x)">{{ shortMoney(x) }}</button></div></label><button>Lưu</button></form><div v-for="b in budgets" :key="b.id" class="progress"><b>{{ b.category }}</b><span>{{ displayMoney(b.spent) }}/{{ displayMoney(b.limit) }}</span><i :class="b.status" :style="{width: Math.min(100,b.percent)+'%'}"></i></div></article></section>
      <section v-show="activeSection === 'goals'" class="tools-grid"><article class="card"><h2>Mục tiêu tiết kiệm</h2><form @submit.prevent="saveGoal"><input v-model="goal.name" placeholder="Mua laptop"/><label class="money-field"><span>Số tiền mục tiêu</span><input v-model="goal.targetAmount" type="text" inputmode="numeric" placeholder="VD: nhập 2 sẽ gợi ý 20k, 200k, 2tr" @blur="normalizeMoney(goal, 'targetAmount')"/><div class="money-presets"><button v-for="x in moneySuggestions(goal.targetAmount)" :key="x" type="button" @click="setMoney(goal, 'targetAmount', x)">{{ shortMoney(x) }}</button></div></label><label class="money-field"><span>Số tiền đã có</span><input v-model="goal.currentAmount" type="text" inputmode="numeric" placeholder="VD: 5tr" @blur="normalizeMoney(goal, 'currentAmount')"/><div class="money-presets"><button v-for="x in moneySuggestions(goal.currentAmount)" :key="x" type="button" @click="setMoney(goal, 'currentAmount', x)">{{ shortMoney(x) }}</button></div></label><label class="date-field"><span>Ngày hoàn thành mục tiêu</span><input v-model="goal.deadline" type="date"/></label><button>Lưu</button></form><div v-for="g in goals" :key="g.id" class="progress"><b>{{ g.name }}</b><span>Cần tiết kiệm/tháng: {{ displayMoney(g.monthlyNeed) }}</span><i class="ok" :style="{width: Math.min(100,g.percent)+'%'}"></i></div></article></section>
      <section v-show="activeSection === 'debts'" class="tools-grid"><article class="card"><h2>Nợ & cho vay</h2><form @submit.prevent="saveDebt"><select v-model="debt.kind"><option value="borrow">Tôi nợ</option><option value="lend">Người khác nợ tôi</option></select><input v-model="debt.person" placeholder="Tên người liên quan"/><label class="money-field"><span>Số tiền nợ / cho vay</span><input v-model="debt.amount" type="text" inputmode="numeric" placeholder="Nhập 1 → chọn 10k / 100k / 1tr" @blur="normalizeMoney(debt, 'amount')"/><div class="money-presets"><button v-for="x in moneySuggestions(debt.amount)" :key="x" type="button" @click="setMoney(debt, 'amount', x)">{{ shortMoney(x) }}</button></div></label><label class="date-field"><span>Ngày hẹn trả / thu nợ</span><input v-model="debt.dueDate" type="date"/></label><select v-model="debt.walletId"><option v-for="w in coreWallets" :key="w.id" :value="w.id">{{ w.name }}</option></select><textarea v-model="debt.note" placeholder="Ghi chú"></textarea><button>Lưu khoản nợ</button></form><div v-for="d in debts" :key="d.id" class="debt-row">
  <div>
    <b>{{ d.kind === 'borrow' ? 'Tôi nợ' : 'Cho vay' }} - {{ d.person }}</b>
    <span>{{ displayMoney(d.amount) }} • hạn {{ d.dueDate || 'không có' }} • {{ d.status === 'done' ? 'Đã hoàn thành' : 'Đang theo dõi' }}</span>
  </div>
  <div class="debt-actions">
    <button v-if="d.status !== 'done'" type="button" class="primary" @click="completeDebt(d)">{{ d.kind === 'borrow' ? 'Đã trả nợ' : 'Đã thu nợ' }}</button>
    <button type="button" class="danger" @click="removeDebt(d)">Xóa</button>
  </div>
</div></article></section>
      <section v-show="activeSection === 'settings'" class="tools-grid"><article class="card"><h2>Nhắc nhở Telegram</h2><p class="muted">Chọn giờ nhắc mỗi ngày để bot Telegram nhắn bạn ghi lại chi tiêu.</p><form @submit.prevent="saveTelegramReminder"><label class="chip"><span>Bật nhắc nhở hằng ngày</span><input v-model="telegramReminder.enabled" type="checkbox" /></label><label class="date-field"><span>Giờ nhắc mỗi ngày</span><input v-model="telegramReminder.time" type="time" /></label><input v-model.trim="telegramReminder.telegramChatId" placeholder="Nhập Chat ID, ví dụ: 5687993964"/><div class="actions"><button>Lưu nhắc nhở Telegram</button><button type="button" class="primary" @click="testTelegramReminder">Gửi thử bot</button></div></form><div class="guide-mini"><b>Cách lấy Chat ID:</b><ol><li>Tạo bot bằng @BotFather và cấu hình TELEGRAM_BOT_TOKEN trên server.</li><li>Nhắn /start cho bot của bạn.</li><li>Nhập Chat ID của bạn, không nhập @username của bot.</li><li>Bấm “Gửi thử bot” để kiểm tra ngay.</li></ol></div></article><article class="card"><h2>Cài đặt & riêng tư</h2><p class="muted">Import CSV mẫu: Ngày, Loại, Danh mục, Tên khoản, Số tiền, Ghi chú, Loại chi phí.</p><input type="file" accept=".csv" @change="importCsv"/><label class="chip"><span>Nhắc ghi chép hằng ngày lúc 21:00</span><input v-model="reminder" type="checkbox" /></label><button class="danger" @click="deleteAccount">Xóa tài khoản và toàn bộ dữ liệu</button></article></section>
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
  {id:'categories',icon:'🏷️',label:'Danh mục'},
  {id:'wallets',icon:'👛',label:'Ví'},
  {id:'budgets',icon:'🎯',label:'Ngân sách'},
  {id:'goals',icon:'🐷',label:'Tiết kiệm'},
  {id:'debts',icon:'🤝',label:'Nợ vay'},
  {id:'settings',icon:'⚙️',label:'Cài đặt'}
]
const sectionHelp={
  dashboard:{title:'Tổng quan tài chính',desc:'Xem nhanh sức khỏe tài chính trong tháng đang chọn.',tips:['Số dư ví là tổng tiền hiện có trong tất cả ví.','Biểu đồ danh mục cho biết tiền chi nhiều nhất ở đâu.','Bấm “Ẩn số dư” khi dùng app ở nơi đông người.']},
  transactions:{title:'Giao dịch thu / chi',desc:'Chỉ dùng để thêm hoặc sửa khoản thu chi hằng ngày.',tips:['Nhập tên khoản, số tiền, ngày giao dịch và chọn đúng ví.','Chọn danh mục để báo cáo và ngân sách tính chính xác.','Nếu cần xem lịch sử hoặc xuất Excel, hãy sang mục Truy vấn.']},
  transactionQuery:{title:'Truy vấn giao dịch',desc:'Xem lại lịch sử giao dịch theo ngày và xuất Excel.',tips:['Chọn Từ ngày và Đến ngày rồi bấm Truy vấn.','Có thể lọc tiền vào/tiền ra hoặc tìm theo từ khóa.','Danh sách được gom theo từng ngày giống app ngân hàng.']},
  categories:{title:'Danh mục',desc:'Tạo nhóm thu chi như Ăn uống, Di chuyển, Lương, Đầu tư.',tips:['Danh mục giúp báo cáo rõ tiền đi vào đâu.','Đổi tên danh mục sẽ cập nhật lại giao dịch cũ cùng tên.']},
  wallets:{title:'Ví và tài khoản',desc:'Quản lý tiền trong 2 nơi chính: Ví tiền mặt và Ngân hàng.',tips:['Ứng dụng chỉ dùng 2 ví cố định: Ví tiền mặt và Ngân hàng.','Mỗi giao dịch cần chọn tiền đi từ Ví tiền mặt hay Ngân hàng.','Bạn có thể tự đặt mức tiền tối thiểu để cảnh báo số dư thấp.']},
  budgets:{title:'Ngân sách',desc:'Đặt hạn mức chi tiêu theo danh mục trong tháng.',tips:['Thanh xanh là an toàn, vàng là gần 80%, đỏ là vượt hoặc chạm hạn mức.','Dùng ngân sách để kiểm soát các khoản dễ vượt như ăn uống, mua sắm.']},
  goals:{title:'Mục tiêu tiết kiệm',desc:'Theo dõi mục tiêu lớn như mua laptop, du lịch, quỹ khẩn cấp.',tips:['Nhập số tiền mục tiêu, số đã có và hạn hoàn thành.','Hệ thống tính số tiền cần tiết kiệm mỗi tháng.']},
  debts:{title:'Nợ và cho vay',desc:'Theo dõi bạn nợ ai hoặc ai đang nợ bạn.',tips:['“Tôi nợ” sẽ trừ tiền ví khi bấm Đã trả nợ.','“Người khác nợ tôi” sẽ cộng tiền ví khi bấm Đã thu nợ.','Luôn chọn đúng ví dùng để trả hoặc nhận tiền.']},
  settings:{title:'Cài đặt & riêng tư',desc:'Nhập dữ liệu, bật nhắc nhở và quản lý quyền riêng tư.',tips:['Import CSV dùng để chuyển dữ liệu từ file cũ vào app.','Xóa tài khoản sẽ xóa toàn bộ dữ liệu và không thể hoàn tác.']}
}
const user=ref(null), authMode=ref('login'), activeSection=ref(localStorage.getItem('activeSection')||'dashboard'), chartPeriod=ref(localStorage.getItem('chartPeriod')||'month'), chartDate=ref(today), chartMonth=ref(currentMonth), chartYear=ref(new Date().getFullYear()), dark=ref(localStorage.getItem('dark')==='1'), hideMoney=ref(false), lowBalanceLimit=ref(Number(localStorage.getItem('lowBalanceLimit')||200000)), reminder=ref(localStorage.getItem('reminder')==='1'), error=ref(''), editingId=ref(null), mobileMenuOpen=ref(false)
if(!tabs.some(t=>t.id===activeSection.value)) activeSection.value='dashboard'
const expenses=ref([]), categories=ref([]), wallets=ref([]), budgets=ref([]), goals=ref([]), debts=ref([])
const summary=ref({income:0,expense:0,balance:0,byCategory:{},dailyIncome:{},dailyExpense:{}})
const auth=reactive({name:'',email:'demo@example.com',password:'demo123456'}), filters=reactive({month:'',search:'',type:'',from:'',to:'',costKind:''})
const blank=()=>({title:'',amount:null,category:'',type:'expense',date:today,note:'',walletId:0,tags:[],costKind:'variable'}); const form=reactive(blank())
const newCategory=reactive({name:'',type:'expense'}), newWallet=reactive({name:'',balance:null}), budget=reactive({category:'',month:currentMonth,limit:null}), goal=reactive({name:'',targetAmount:null,currentAmount:null,deadline:''}), debt=reactive({kind:'borrow',person:'',amount:null,dueDate:'',note:'',walletId:0}), telegramReminder=reactive({enabled:false,time:'21:00',telegramChatId:''})
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
function saveChartPeriod(){localStorage.setItem('chartPeriod', chartPeriod.value)}
function syncChartDateFromMonth(){chartDate.value=(chartMonth.value||currentMonth)+'-01'; loadData()}
function syncChartDateFromYear(){chartDate.value=(chartYear.value||new Date().getFullYear())+'-01-01'; loadData()}
const coreWallets=computed(()=>wallets.value.filter(w=>['Ví tiền mặt','Ngân hàng'].includes(w.name)))
const walletBalance=computed(()=>coreWallets.value.reduce((s,w)=>s+w.balance,0)), lowWallets=computed(()=>coreWallets.value.filter(w=>w.balance < lowBalanceLimit.value)), exportHref=computed(()=>api.exportUrl(filters))
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
async function submitAuth(){try{error.value=''; const res=authMode.value==='login'?await api.login(auth):await api.register(auth); setToken(res.token); user.value=res.user; await loadData()}catch(e){error.value=e.message}}
async function logout(){await api.logout().catch(()=>{}); setToken(''); user.value=null}
async function loadData(){try{const [ex,sm,cs,ws,bs,gs,ds,rem]=await Promise.all([api.expenses(filters),api.summary(dashboardQuery()),api.categories(),api.wallets(),api.budgets(budget.month || currentMonth),api.goals(),api.debts(),api.reminder()]); expenses.value=ex; summary.value=sm; categories.value=cs; wallets.value=ws; budgets.value=bs; goals.value=gs; debts.value=ds; Object.assign(telegramReminder,{enabled:!!rem.enabled,time:rem.time||'21:00',telegramChatId:rem.telegramChatId||''}); if(!form.walletId&&coreWallets.value[0])form.walletId=coreWallets.value[0].id; if(!debt.walletId&&coreWallets.value[0])debt.walletId=coreWallets.value[0].id}catch(e){error.value=e.message}}
function debouncedLoad(){clearTimeout(timer); timer=setTimeout(loadData,250)}
function toggleTheme(){dark.value=!dark.value; localStorage.setItem('dark', dark.value?'1':'0')}
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
async function submitExpense(){try{normalizeMoney(form,'amount'); form.tags=[]; editingId.value?await api.updateExpense(editingId.value,form):await api.createExpense(form); resetForm(); await loadData()}catch(e){error.value=e.message}}
function editItem(e){Object.assign(form,{...e,tags:[]}); editingId.value=e.id; scrollTo({top:0,behavior:'smooth'})}
async function removeItem(id){if(confirm('Xóa giao dịch này?')){await api.deleteExpense(id); await loadData()}}
function resetForm(){Object.assign(form,blank()); form.walletId=coreWallets.value[0]?.id||0; editingId.value=null}
async function saveCategory(){await api.createCategory(newCategory); newCategory.name=''; await loadData()}
async function renameCategory(c){const name=prompt('Tên danh mục mới',c.name); if(name){await api.updateCategory(c.id,{name,type:c.type}); await loadData()}}
async function removeCategory(c){if(confirm(`Xóa danh mục "${c.name}"? Các giao dịch dùng danh mục này vẫn giữ nguyên tên.`)){await api.deleteCategory(c.id); await loadData()}}
async function saveWallet(){await api.createWallet(newWallet); newWallet.name=''; newWallet.balance=null; await loadData()}
async function saveBudget(){normalizeMoney(budget,'limit'); await api.saveBudget({...budget,month:budget.month || currentMonth}); budget.limit=null; await loadData()}
async function saveGoal(){normalizeMoney(goal,'targetAmount'); normalizeMoney(goal,'currentAmount'); await api.saveGoal(goal); Object.assign(goal,{name:'',targetAmount:null,currentAmount:null,deadline:''}); await loadData()}
async function saveDebt(){normalizeMoney(debt,'amount'); debt.walletId=debt.walletId||coreWallets.value[0]?.id||0; await api.saveDebt(debt); Object.assign(debt,{kind:'borrow',person:'',amount:null,dueDate:'',note:'',walletId:coreWallets.value[0]?.id||0}); await loadData()}
async function completeDebt(d){await api.completeDebt(d.id,d.walletId); await loadData()}
async function removeDebt(d){
  const note=d.status==='done'?'\nKhoản này đã hoàn thành nên khi xóa hệ thống sẽ hoàn tác lại số dư ví liên quan.':''
  if(confirm(`Xóa khoản nợ/cho vay của ${d.person}?${note}`)){await api.deleteDebt(d.id); await loadData()}
}
async function importCsv(e){const file=e.target.files?.[0]; if(file){const r=await api.importCsv(file); alert(`Đã import ${r.imported} giao dịch`); await loadData()}}
async function saveTelegramReminder(){await api.saveReminder(telegramReminder); alert('Đã lưu nhắc nhở Telegram')}
async function testTelegramReminder(){try{await api.saveReminder(telegramReminder); await api.testReminder(); alert('Đã gửi tin nhắn test. Kiểm tra Telegram nhé!')}catch(e){alert(e.message)}}
async function deleteAccount(){if(confirm('Xóa vĩnh viễn tài khoản và toàn bộ dữ liệu?')){await api.deleteAccount(); setToken(''); user.value=null}}
function displayMoney(v){return hideMoney.value?'******':money(v)}
function money(v){return new Intl.NumberFormat('vi-VN',{style:'currency',currency:'VND'}).format(v||0)}
function plainMoney(v){return new Intl.NumberFormat('vi-VN').format(v||0)}
function formatDate(v){return new Intl.DateTimeFormat('vi-VN').format(new Date(v))}
function formatTime(v){const d=v?new Date(v):new Date(); return d.toLocaleTimeString('vi-VN',{hour:'2-digit',minute:'2-digit'})}
function walletName(id){return wallets.value.find(w=>w.id===id)?.name || 'Ví'}
onMounted(async()=>{try{const me=await api.me(); user.value=me; await loadData()}catch(_){}})
</script>
