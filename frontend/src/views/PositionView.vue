<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getBatchQuotes, searchStocks as searchStocksApi, getStockScores, getPositionHealthScore } from '@/api'
import type { StockScore } from '@/types/stock'
import StockDetailSheet from '@/components/StockDetailSheet.vue'
import StockTag from '@/components/StockTag.vue'

interface Position { id: string; code: string; name: string; quantity: number; costPrice: number; addTime: number }
interface PositionDisplay extends Position {
  currentPrice: number; changePercent: number; marketValue: number; profit: number; profitPercent: number
  signal?: string; signalType?: 'sell' | 'hold' | 'buy'; advice?: string
  totalScore?: number; positionHealthScore?: number; positionHealthLabel?: string
  macdSignal?: string; bollPosition?: string; highlights?: string[]; recommendation?: string
}

const STORAGE_KEY = 'stock_position_list'
const STOP_LOSS_KEY = 'stop_loss_percent'
const positions = ref<PositionDisplay[]>([])
const dialogMode = ref<'add' | 'edit' | null>(null)
const selectedPosition = ref<PositionDisplay | null>(null)
const isDetailSheetOpen = ref(false); const detailCode = ref(''); const detailCostPrice = ref(0)
const form = ref({ code: '', name: '', quantity: '', costPrice: '' })
const editForm = ref({ code: '', name: '', quantity: '', costPrice: '' })
const searchResults = ref<Array<{ code: string; name: string }>>([])
const searchLoading = ref(false); const isTradingSession = ref(false)
let tradingInterval: ReturnType<typeof setInterval> | null = null
let refreshInterval: ReturnType<typeof setInterval> | null = null

const loadPositions = (): Position[] => { try { return JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]') } catch { return [] } }
const savePositions = (list: Position[]) => { localStorage.setItem(STORAGE_KEY, JSON.stringify(list)) }
const getStopLossPercent = (): number => { const v = localStorage.getItem(STOP_LOSS_KEY); return v ? parseFloat(v) : -8 }

const isTradingTime = (): boolean => {
  const n = new Date(); const d = n.getDay(); const h = n.getHours(); const m = n.getMinutes()
  if (d === 0 || d === 6) return false
  if ((h === 9 && m >= 30) || (h >= 10 && h < 12)) return true
  if (h >= 13 && h < 15) return true; return false
}

interface TradeSignal { signal: string; type: 'SELL' | 'BUY' | 'HOLD' | 'WARNING'; level: 'HIGH' | 'MEDIUM' | 'LOW'; title: string; reason: string }
const generateSignal = (price: number, costPrice: number, changePercent: number, stopLoss: number): TradeSignal | null => {
  const pp = ((price - costPrice) / costPrice) * 100
  if (pp <= stopLoss) return { signal:'止损出局', type:'WARNING', level:'HIGH', title:'触发止损', reason:`亏损${pp.toFixed(1)}%，超过${stopLoss}%止损线` }
  if (pp >= 15 && changePercent > 0) return { signal:'建议止盈', type:'SELL', level:'MEDIUM', title:'高位止盈', reason:`盈利${pp.toFixed(1)}%，可分批锁定利润` }
  if (changePercent <= -3) return { signal:'注意风险', type:'WARNING', level:'LOW', title:'注意风险', reason:`今日下跌${changePercent.toFixed(1)}%，关注是否企稳` }
  if (pp > 0) return { signal:'持仓盈利', type:'HOLD', level:'LOW', title:'建议持有', reason:`盈利${pp.toFixed(1)}%，趋势向好` }
  if (pp > -5) return { signal:'轻仓观望', type:'HOLD', level:'LOW', title:'轻仓观望', reason:'小幅亏损，继续观察' }
  return null
}

const loadPositionData = async () => {
  try {
    const stored = loadPositions(); const codes = stored.map(p => p.code)
    if (codes.length > 0) {
      const res = await getBatchQuotes(codes)
      if (res.code === 200) {
        const quotes = (res.data || {}) as Record<string, { price: string; changePercent: string }>
        const stopLoss = getStopLossPercent()
        positions.value = stored.map(pos => {
          const q = quotes[pos.code] || {}; const cp = parseFloat(q.price) || 0; const ch = parseFloat(q.changePercent) || 0
          const mv = cp * pos.quantity; const cv = pos.costPrice * pos.quantity; const pft = mv - cv; const pfp = cv > 0 ? (pft / cv) * 100 : 0
          const sig = generateSignal(cp, pos.costPrice, ch, stopLoss)
          return { ...pos, currentPrice: cp, changePercent: ch, marketValue: mv, profit: pft, profitPercent: pfp, signal: sig?.signal, signalType: sig?.type === 'WARNING' || sig?.type === 'SELL' ? 'sell' : sig?.type === 'BUY' ? 'buy' : 'hold' as const, advice: sig?.reason }
        })
        loadStockScores(codes)
      }
    } else { positions.value = [] }
  } catch (err) { console.error('加载失败:', err) }
}

const loadStockScores = async (codes: string[]) => {
  try {
    const res = await getStockScores(codes)
    if (res.code === 200) {
      const scores = (res.data || {}) as Record<string, StockScore>
      positions.value = positions.value.map(pos => {
        const s = scores[pos.code]
        return s ? { ...pos, totalScore: s.totalScore, macdSignal: s.macdSignal, bollPosition: s.bollPosition, highlights: s.highlights || [], recommendation: s.recommendation } : pos
      })
    }
    for (const pos of positions.value) {
      if (pos.costPrice > 0) {
        try { const h = await getPositionHealthScore(pos.code, pos.costPrice); if (h.code === 200 && h.data) { pos.positionHealthScore = (h.data as StockScore).positionHealthScore; pos.positionHealthLabel = (h.data as StockScore).positionHealthLabel } } catch {}
      }
    }
  } catch (err) { console.error('获取评分失败:', err) }
}

const refreshPositionPrices = async () => {
  if (positions.value.length === 0) return
  try {
    const codes = positions.value.map(p => p.code); const res = await getBatchQuotes(codes)
    if (res.code === 200) {
      const quotes = (res.data || {}) as Record<string, { price: string; changePercent: string }>
      const stopLoss = getStopLossPercent()
      positions.value = positions.value.map(pos => {
        const q = quotes[pos.code] || {}; const cp = parseFloat(q.price) || pos.currentPrice; const ch = parseFloat(q.changePercent) || pos.changePercent
        const mv = cp * pos.quantity; const cv = pos.costPrice * pos.quantity; const pft = mv - cv; const pfp = cv > 0 ? (pft / cv) * 100 : 0
        const sig = generateSignal(cp, pos.costPrice, ch, stopLoss)
        return { ...pos, currentPrice: cp, changePercent: ch, marketValue: mv, profit: pft, profitPercent: pfp, signal: sig?.signal, signalType: sig?.type === 'WARNING' || sig?.type === 'SELL' ? 'sell' : sig?.type === 'BUY' ? 'buy' : 'hold' as const, advice: sig?.reason }
      })
    }
  } catch (err) { console.error('刷新价格失败:', err) }
}

const handleSearch = async (keyword: string) => {
  if (keyword.length < 1) { searchResults.value = []; return }
  searchLoading.value = true
  try { const res = await searchStocksApi(keyword); searchResults.value = res.code === 200 ? (res.data || []) : [] } finally { searchLoading.value = false }
}
const selectStock = (s: { code: string; name: string }) => { form.value = { ...form.value, code: s.code, name: s.name }; searchResults.value = []; dialogMode.value = 'add' }
const addPosition = () => {
  const q = parseInt(form.value.quantity); const c = parseFloat(form.value.costPrice)
  if (!form.value.code || Number.isNaN(q) || q <= 0 || Number.isNaN(c) || c <= 0) return
  const np: Position = { id: `pos_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`, code: form.value.code, name: form.value.name, quantity: q, costPrice: c, addTime: Date.now() }
  savePositions([...loadPositions(), np]); dialogMode.value = null; form.value = { code: '', name: '', quantity: '', costPrice: '' }; loadPositionData()
}
const openEditDialog = (pos: PositionDisplay) => { selectedPosition.value = pos; editForm.value = { code: pos.code, name: pos.name, quantity: pos.quantity.toString(), costPrice: pos.costPrice.toString() }; dialogMode.value = 'edit' }
const saveEdit = () => {
  if (!selectedPosition.value) return
  const q = parseInt(editForm.value.quantity); const c = parseFloat(editForm.value.costPrice)
  if (Number.isNaN(q) || q <= 0 || Number.isNaN(c) || c <= 0) return
  const updated: Position = { ...selectedPosition.value, quantity: q, costPrice: c }
  const stored = loadPositions(); const idx = stored.findIndex(p => p.id === selectedPosition.value!.id)
  if (idx >= 0) { stored[idx] = updated; savePositions(stored) }
  positions.value = positions.value.map(p => p.id === selectedPosition.value!.id ? { ...p, quantity: q, costPrice: c, marketValue: p.currentPrice * q, profit: p.currentPrice * q - c * q, profitPercent: c * q > 0 ? (p.currentPrice * q - c * q) / (c * q) * 100 : 0 } : p)
  dialogMode.value = null
}
const handleDeletePosition = (pos: PositionDisplay) => {
  if (!confirm(`确定要删除 ${pos.name} 的持仓记录吗？`)) return
  savePositions(loadPositions().filter(p => p.id !== pos.id)); positions.value = positions.value.filter(p => p.id !== pos.id)
}
const openDetailSheet = (pos: PositionDisplay) => { detailCode.value = pos.code; detailCostPrice.value = pos.costPrice; isDetailSheetOpen.value = true }
const totalProfit = computed(() => positions.value.reduce((s, p) => s + p.profit, 0))
const totalValue = computed(() => positions.value.reduce((s, p) => s + p.marketValue, 0))
const totalCost = computed(() => positions.value.reduce((s, p) => s + p.costPrice * p.quantity, 0))
const totalProfitPercent = computed(() => totalCost.value > 0 ? (totalProfit.value / totalCost.value) * 100 : 0)

onMounted(() => {
  loadPositionData(); isTradingSession.value = isTradingTime()
  tradingInterval = setInterval(() => { isTradingSession.value = isTradingTime() }, 60000)
  refreshInterval = setInterval(() => { if (isTradingSession.value && positions.value.length > 0) refreshPositionPrices() }, 5000)
})
onUnmounted(() => { if (tradingInterval) clearInterval(tradingInterval); if (refreshInterval) clearInterval(refreshInterval) })
</script>

<template>
  <div class="space-y-5" style="max-width: 700px; margin: 0 auto;">
    <!-- 顶部统计 — 紧凑行 -->
    <div class="card p-5 flex items-center justify-between">
      <div class="flex items-center gap-6">
        <div>
          <div class="text-text-tertiary text-[10px] tracking-wide">总市值</div>
          <div class="num-font text-text-primary text-xl font-semibold mt-0.5">¥{{ totalValue.toFixed(2) }}</div>
        </div>
        <div class="w-px h-8 bg-black/6" />
        <div>
          <div class="text-text-tertiary text-[10px] tracking-wide">持仓成本</div>
          <div class="num-font text-text-secondary text-sm mt-0.5">¥{{ totalCost.toFixed(2) }}</div>
        </div>
        <div class="w-px h-8 bg-black/6" />
        <div>
          <div class="text-text-tertiary text-[10px] tracking-wide">总盈亏</div>
          <div :class="['num-font text-sm font-semibold mt-0.5', totalProfit >= 0 ? 'text-up' : 'text-down']">{{ totalProfit >= 0 ? '+' : '' }}{{ totalProfit.toFixed(2) }} ({{ totalProfitPercent >= 0 ? '+' : '' }}{{ totalProfitPercent.toFixed(2) }}%)</div>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-text-tertiary text-[10px] bg-white/50 px-2 py-0.5 rounded-full">{{ isTradingSession ? '交易中' : '已收盘' }}</span>
        <span v-if="isTradingSession" class="w-1.5 h-1.5 rounded-full bg-up breathe-dot inline-block" />
        <button class="btn btn-primary text-xs px-4 py-2 flex items-center gap-1.5" @click="dialogMode = 'add'">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加
        </button>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="positions.length === 0" class="card p-8 flex flex-col items-center gap-3 text-center"
      style="border: 1.5px dashed rgba(0,0,0,0.06); background: rgba(255,255,255,0.45);">
      <svg class="w-10 h-10 text-text-tertiary/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
      <div class="text-text-secondary text-sm">暂无持仓记录</div>
      <button class="btn btn-primary px-5 py-2 text-sm" @click="dialogMode = 'add'">+ 添加第一笔持仓</button>
    </div>

    <!-- 持仓卡片 — 单列满宽 -->
    <div v-if="positions.length > 0" class="space-y-3">
      <div v-for="(pos, idx) in positions" :key="pos.id"
        class="card p-5 spring-up" :style="{ animationDelay: `${idx * 0.04}s` }">
        <div class="flex items-start justify-between cursor-pointer" @click="openDetailSheet(pos)">
          <div class="min-w-0 flex-1 pr-3">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-text-primary font-semibold text-base">{{ pos.name }}</span>
              <StockTag v-if="pos.signal && pos.signalType" type="signalType" :text="pos.signalType">{{ pos.signal }}</StockTag>
            </div>
            <div class="text-text-tertiary text-xs mt-0.5">{{ pos.code }}</div>
          </div>
          <div class="text-right shrink-0 flex items-center gap-4">
            <div>
              <div class="num-font text-text-primary text-lg font-semibold">¥{{ pos.currentPrice.toFixed(2) }}</div>
              <div :class="['num-font text-sm flex items-center justify-end gap-1', pos.changePercent >= 0 ? 'text-up' : 'text-down']">
                <svg v-if="pos.changePercent >= 0" class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><polyline points="18 15 12 9 6 15"/></svg>
                <svg v-else class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><polyline points="6 9 12 15 18 9"/></svg>
                {{ pos.changePercent >= 0 ? '+' : '' }}{{ pos.changePercent.toFixed(2) }}%
              </div>
            </div>
            <div class="text-right">
              <div class="text-text-tertiary text-[10px]">盈亏</div>
              <div :class="['num-font text-sm font-semibold mt-0.5', pos.profit >= 0 ? 'text-up' : 'text-down']">{{ pos.profit >= 0 ? '+' : '' }}{{ pos.profit.toFixed(2) }}</div>
              <div :class="['num-font text-[11px]', pos.profitPercent >= 0 ? 'text-up' : 'text-down']">{{ pos.profitPercent >= 0 ? '+' : '' }}{{ pos.profitPercent.toFixed(2) }}%</div>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-3 mt-3 flex-wrap">
          <div class="flex items-center gap-1.5 text-xs text-text-tertiary"><span>持仓</span><span class="num-font text-text-secondary font-medium">{{ pos.quantity }}股</span></div>
          <span class="text-text-tertiary/30">·</span>
          <div class="flex items-center gap-1.5 text-xs text-text-tertiary"><span>成本</span><span class="num-font text-text-secondary font-medium">¥{{ pos.costPrice.toFixed(2) }}</span></div>
          <span class="text-text-tertiary/30">·</span>
          <div class="flex items-center gap-1.5 text-xs text-text-tertiary"><span>市值</span><span class="num-font text-text-secondary font-medium">¥{{ pos.marketValue.toFixed(2) }}</span></div>
          <StockTag v-if="pos.totalScore !== undefined" type="score" :score="pos.totalScore" :text="`评分 ${pos.totalScore.toFixed(0)}`" />
          <StockTag v-if="pos.macdSignal" type="macdSignal" :text="pos.macdSignal" />
        </div>

        <!-- 建议 -->
        <div v-if="pos.advice" class="mt-2 bg-white/70 rounded-xl px-3.5 py-2.5 flex items-start gap-2">
          <svg class="w-3.5 h-3.5 text-text-tertiary shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
          <span class="text-text-tertiary text-xs leading-relaxed">{{ pos.advice }}</span>
        </div>

        <!-- 按钮 -->
        <div class="flex gap-3 mt-3">
          <button class="flex-1 py-2 rounded-2xl bg-black/5 text-sm text-text-secondary hover:bg-black/10 transition-all" @click="openEditDialog(pos)"><svg class="w-3.5 h-3.5 inline mr-1 align-middle" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>调整</button>
          <button class="flex-[2] py-2 rounded-2xl btn-primary text-sm font-medium" @click="openDetailSheet(pos)"><svg class="w-3.5 h-3.5 inline mr-1 align-middle" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>详情与止盈</button>
          <button class="flex-1 py-2 rounded-2xl bg-black/5 text-sm text-text-tertiary hover:text-up transition-all" @click="handleDeletePosition(pos)"><svg class="w-3.5 h-3.5 inline mr-1 align-middle" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg></button>
        </div>
      </div>
    </div>

    <!-- ═══ 统一弹窗（添加/编辑/搜索 合并） ═══ -->
    <Teleport to="body">
      <Transition name="drawer">
        <div v-if="dialogMode" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center" @click.self="dialogMode = null">
          <div class="absolute inset-0 modal-overlay" @click="dialogMode = null" />
          <div class="card w-full sm:max-w-md rounded-t-3xl sm:rounded-3xl p-7 max-h-[85vh] overflow-y-auto shadow-2xl relative z-10 !rounded-b-none sm:!rounded-b-3xl">
            <!-- 编辑模式 -->
            <template v-if="dialogMode === 'edit'">
              <h3 class="text-text-primary text-lg font-semibold mb-5">调整持仓</h3>
              <div class="space-y-5">
                <div><label class="text-text-secondary text-sm mb-1.5 block">股票</label><div class="text-text-primary text-sm px-4 py-3 bg-black/4 rounded-2xl">{{ editForm.name }} ({{ editForm.code }})</div></div>
                <div><label class="text-text-secondary text-sm mb-1.5 block">数量（股）</label><input v-model="editForm.quantity" type="number" placeholder="输入数量" class="input-organic w-full text-sm" /></div>
                <div><label class="text-text-secondary text-sm mb-1.5 block">成本价（元/股）</label><input v-model="editForm.costPrice" type="number" step="0.01" placeholder="输入成本价" class="input-organic w-full text-sm" /></div>
              </div>
              <div class="flex gap-3 mt-6">
                <button class="flex-1 py-2.5 rounded-2xl bg-black/5 text-sm text-text-secondary hover:bg-black/10 transition-all" @click="dialogMode = null">取消</button>
                <button class="flex-1 py-2.5 rounded-2xl btn-primary text-sm font-medium" @click="saveEdit">保存</button>
              </div>
            </template>

            <!-- 添加模式（内置搜索） -->
            <template v-else>
              <h3 class="text-text-primary text-lg font-semibold mb-5">添加持仓</h3>
              <div class="space-y-5">
                <div>
                  <label class="text-text-secondary text-sm mb-1.5 block">股票</label>
                  <div v-if="!form.name" class="space-y-2">
                    <input v-model="form.code" placeholder="输入代码或名称搜索" class="input-organic w-full text-sm" @input="handleSearch(form.code)" />
                    <div v-if="searchResults.length > 0" class="max-h-40 overflow-y-auto space-y-0.5 bg-white/80 rounded-2xl p-1">
                      <div v-for="s in searchResults" :key="s.code" class="px-3 py-2 rounded-xl hover:bg-black/5 cursor-pointer transition-all text-sm flex items-center justify-between" @click="selectStock(s)">
                        <span>{{ s.name }}</span>
                        <span class="text-text-tertiary text-xs">{{ s.code }}</span>
                      </div>
                    </div>
                    <div v-else-if="searchLoading" class="text-center text-text-tertiary text-xs py-2">搜索中...</div>
                    <div v-else-if="form.code.length > 1" class="text-center text-text-tertiary text-xs py-2">无匹配结果</div>
                    <div v-else class="text-center text-text-tertiary text-xs py-2">输入代码或名称搜索股票</div>
                  </div>
                  <div v-else class="bg-black/4 rounded-2xl px-4 py-3.5 flex items-center justify-between">
                    <span class="text-sm text-text-primary">{{ form.name }} <span class="text-text-tertiary">({{ form.code }})</span></span>
                    <button class="text-xs text-text-tertiary hover:text-text-secondary" @click="form = { code: '', name: '', quantity: '', costPrice: '' }; searchResults = []">更改</button>
                  </div>
                </div>
                <div><label class="text-text-secondary text-sm mb-1.5 block">数量（股）</label><input v-model="form.quantity" type="number" placeholder="输入持仓数量" class="input-organic w-full text-sm" :disabled="!form.name" /></div>
                <div><label class="text-text-secondary text-sm mb-1.5 block">成本价（元/股）</label><input v-model="form.costPrice" type="number" step="0.01" placeholder="输入成本价" class="input-organic w-full text-sm" :disabled="!form.name" /></div>
              </div>
              <div class="flex gap-3 mt-6">
                <button class="flex-1 py-2.5 rounded-2xl bg-black/5 text-sm text-text-secondary hover:bg-black/10 transition-all" @click="dialogMode = null">取消</button>
                <button class="flex-1 py-2.5 rounded-2xl btn-primary text-sm font-medium" :disabled="!form.name" @click="addPosition">确认添加</button>
              </div>
            </template>
          </div>
        </div>
      </Transition>
    </Teleport>

    <StockDetailSheet v-model:open="isDetailSheetOpen" :code="detailCode" :cost-price="detailCostPrice" />
  </div>
</template>
