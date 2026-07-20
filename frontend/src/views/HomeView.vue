<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getMarketStatus, scanAllAShares } from '@/api'
import type { MarketData, IndexData, AllStockScanResult, RankStockItem } from '@/types/stock'
import StockDetailSheet from '@/components/StockDetailSheet.vue'
import { getRecommendationStyle } from '@/utils/tagColors'
import { useRouter } from 'vue-router'

interface IndexCard { name: string; key: string; short: string }
const ALL_INDICES: IndexCard[] = [
  { name: '上证指数', key: 'shanghai', short: 'SH' },
  { name: '深证成指', key: 'shenzhen', short: 'SZ' },
  { name: '创业板指', key: 'chinext', short: 'CY' },
  { name: '科创50', key: 'star50', short: 'KC' },
  { name: '北证50', key: 'beishang50', short: 'BZ' },
]

const router = useRouter()
const rawMarketData = ref<MarketData | null>(null)
const indicesLoading = ref(true)
const scanResult = ref<AllStockScanResult | null>(null)
const topStocks = ref<RankStockItem[]>([])
const loading = ref(false)
const detailOpen = ref(false)
const selectedCode = ref('')
const marketRefreshTimer = ref<ReturnType<typeof setInterval> | null>(null)

onMounted(() => {
  loadMarketData()
  handleScanAll()
  marketRefreshTimer.value = setInterval(refreshMarketData, 1000)
})

onUnmounted(() => {
  if (marketRefreshTimer.value) { clearInterval(marketRefreshTimer.value); marketRefreshTimer.value = null }
})

const loadMarketData = async () => {
  indicesLoading.value = true
  try {
    const res = await getMarketStatus()
    if (res.code === 200 && res.data) rawMarketData.value = res.data as MarketData
  } catch {} finally { indicesLoading.value = false }
}

const refreshMarketData = async () => {
  try {
    const res = await getMarketStatus()
    if (res.code === 200 && res.data) rawMarketData.value = res.data as MarketData
  } catch {}
}

function getIdx(key: string): IndexData | null {
  return rawMarketData.value?.indices?.[key as keyof MarketData['indices']] ?? null
}

const handleScanAll = async () => {
  loading.value = true
  try {
    const res = await scanAllAShares()
    if (res.code === 200 && res.data) {
      scanResult.value = res.data as AllStockScanResult
      topStocks.value = (res.data as AllStockScanResult).topList || []
    }
  } catch {} finally { loading.value = false }
}

const openDetail = (code: string) => { selectedCode.value = code; detailOpen.value = true }
const goToRank = () => router.push('/rank')
const top10 = computed(() => topStocks.value.slice(0, 10))

// ─── 上证（Hero）额外数据 —──
const shIdx = computed(() => getIdx('shanghai'))
const hePositive = computed(() => (shIdx.value?.changePercent ?? 0) >= 0)

// 价格范围条位置
const rangePos = computed(() => {
  const d = shIdx.value
  if (!d?.high || !d?.low || !d?.price || d.high === d.low) return 50
  return ((d.price - d.low) / (d.high - d.low)) * 100
})

// 副卡涨跌信息
function indexInfo(key: string) {
  const d = getIdx(key)
  if (!d) return { cls: '', icon: '—', text: '--', changeText: '' }
  const pos = d.changePercent >= 0
  return {
    cls: pos ? 'text-up' : 'text-down',
    icon: pos ? '▲' : '▼',
    text: `${pos ? '+' : ''}${d.changePercent.toFixed(2)}%`,
    changeText: `${pos ? '+' : ''}${d.change.toFixed(2)}`,
  }
}
</script>

<template>
  <div class="space-y-4 flex flex-col" style="min-height: calc(100vh - 8rem);">
    <!-- ═══ 指数条：5个指数一行横向排列 ═══ -->
    <div v-if="indicesLoading" class="flex gap-3">
      <div v-for="n in 5" :key="n" class="flex-1 card p-4 space-y-2 animate-pulse">
        <div class="h-3 w-14 skeleton-shimmer rounded-full" />
        <div class="h-6 w-20 skeleton-shimmer rounded-xl" />
      </div>
    </div>

    <div v-else class="flex gap-3 overflow-x-auto pb-1 spring-up">
      <!-- 上证（略宽，带价格范围条） -->
      <div class="min-w-[220px] flex-[1.6] rounded-2xl p-4 transition-all duration-300"
        :class="hePositive
          ? 'bg-gradient-to-br from-white to-up-bg/30 shadow-[0_2px_12px_rgba(212,131,122,0.06)]'
          : 'bg-gradient-to-br from-white to-down-bg/30 shadow-[0_2px_12px_rgba(125,184,154,0.06)]'"
        style="border:1px solid rgba(255,255,255,0.5);">
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs font-medium text-text-tertiary tracking-wide">上证指数</span>
          <span class="flex items-center gap-1 text-[9px] text-text-secondary bg-white/50 px-1.5 py-0.5 rounded-full">
            <span class="w-1 h-1 rounded-full bg-earth-sage breathe-dot inline-block" /> 实时
          </span>
        </div>
        <div class="flex items-baseline gap-2.5 mb-1.5">
          <span class="text-[26px] font-bold text-text-primary num-font tracking-tight leading-none">
            {{ shIdx?.price.toFixed(2) ?? '--' }}
          </span>
          <span :class="['text-sm font-semibold num-font', hePositive ? 'text-up' : 'text-down']">
            {{ hePositive ? '▲' : '▼' }} {{ hePositive ? '+' : '' }}{{ shIdx?.changePercent.toFixed(2) ?? '0.00' }}%
          </span>
        </div>
        <div class="flex items-center gap-2 text-[11px] text-text-tertiary mb-2.5">
          <span>开 {{ shIdx?.open?.toFixed(2) ?? '--' }}</span>
          <span>高 {{ shIdx?.high?.toFixed(2) ?? '--' }}</span>
          <span>低 {{ shIdx?.low?.toFixed(2) ?? '--' }}</span>
          <span class="ml-auto num-font">{{ shIdx?.volume ? (shIdx.volume / 1e8).toFixed(1) + '亿' : '--' }}</span>
        </div>
        <!-- 价格范围条 -->
        <div class="relative h-1 rounded-full overflow-hidden"
          :class="hePositive ? 'bg-up-bg/40' : 'bg-down-bg/40'">
          <div class="absolute top-1/2 -translate-y-1/2 w-2.5 h-2.5 rounded-full shadow-sm transition-all duration-500"
            :class="hePositive ? 'bg-up' : 'bg-down'"
            :style="{ left: `calc(${rangePos}% - 5px)` }" />
        </div>
      </div>

      <!-- 其他4个指数（紧凑） -->
      <div v-for="(card, idx) in ALL_INDICES.slice(1)" :key="card.key"
        class="min-w-[130px] flex-1 rounded-2xl p-3.5 transition-all duration-300 hover-lift spring-up"
        :style="{
          animationDelay: `${(idx) * 0.06}s`,
          background: indexInfo(card.key).cls.includes('up')
            ? 'linear-gradient(135deg, rgba(212,131,122,0.06), #fff 60%)'
            : 'linear-gradient(135deg, rgba(125,184,154,0.06), #fff 60%)',
          border: '1px solid rgba(255,255,255,0.5)',
        }">
        <div class="text-[10px] font-medium text-text-tertiary mb-1.5 tracking-wide">{{ card.name }}</div>
        <template v-if="getIdx(card.key)">
          <div class="text-lg font-bold text-text-primary num-font tracking-tight mb-0.5">
            {{ getIdx(card.key)!.price.toFixed(2) }}
          </div>
          <div :class="['text-xs font-semibold num-font', indexInfo(card.key).cls]">
            {{ indexInfo(card.key).icon }} {{ indexInfo(card.key).text }}
            <span class="text-[10px] opacity-60 ml-1">{{ indexInfo(card.key).changeText }}</span>
          </div>
        </template>
        <template v-else>
          <div class="text-lg font-bold text-text-tertiary/40">--</div>
        </template>
      </div>
    </div>

    <!-- ═══ 操作栏 ═══ -->
    <div class="glass rounded-2xl px-5 py-3.5 flex items-center gap-4 shrink-0">
      <button class="btn btn-primary text-sm px-5 py-2 flex items-center gap-2 disabled:opacity-50 shrink-0"
        :disabled="loading" @click="handleScanAll">
        <svg v-if="loading" class="animate-spin" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10" stroke-dasharray="30 70" stroke-linecap="round"/></svg>
        <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
        {{ loading ? '扫描中...' : '全A股扫描' }}
      </button>
      <span v-if="scanResult" class="text-xs text-text-tertiary bg-white/40 px-2.5 py-1 rounded-full whitespace-nowrap">
        {{ scanResult.totalStocks }} 支 · {{ (scanResult.costMs / 1000).toFixed(1) }}s
      </span>
    </div>

    <!-- ═══ Top 10 排名（桌面临表，移动端卡片） ═══ -->
    <div v-if="top10.length > 0" class="flex-1 card overflow-hidden" style="min-height: 0;">
      <!-- 表头行 -->
      <div class="flex items-center justify-between px-4 md:px-5 pt-4 pb-1">
        <span class="text-xs font-semibold text-text-secondary tracking-wide">综合评分 Top 10</span>
        <button class="text-xs text-earth-sage font-medium hover:text-earth-sage/80 transition-all" @click="goToRank">完整排名 →</button>
      </div>

      <!-- Desktop: table -->
      <div class="hidden md:block h-full overflow-y-auto">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr>
                <th class="text-left text-xs font-medium text-text-tertiary py-3 px-5 w-10 sticky top-0 bg-white/90 backdrop-blur-md z-10">#</th>
                <th class="text-left text-xs font-medium text-text-tertiary py-3 px-5 sticky top-0 bg-white/90 backdrop-blur-md z-10">股票</th>
                <th class="text-right text-xs font-medium text-text-tertiary py-3 px-5 num-font sticky top-0 bg-white/90 backdrop-blur-md z-10">现价</th>
                <th class="text-right text-xs font-medium text-text-tertiary py-3 px-5 num-font sticky top-0 bg-white/90 backdrop-blur-md z-10">涨跌幅</th>
                <th class="text-right text-xs font-medium text-text-tertiary py-3 px-5 sticky top-0 bg-white/90 backdrop-blur-md z-10">综合评分</th>
                <th class="text-center text-xs font-medium text-text-tertiary py-3 px-5 sticky top-0 bg-white/90 backdrop-blur-md z-10">建议</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(stock, index) in top10" :key="stock.code"
                class="cursor-pointer transition-all duration-200 hover:bg-black/4 spring-up"
                :style="{ animationDelay: `${index * 0.04}s` }"
                @click="openDetail(stock.code)">
                <td class="py-2.5 px-5">
                  <span class="inline-flex items-center justify-center w-6 h-6 rounded-xl text-xs font-semibold"
                    :class="index === 0 ? 'bg-earth-sage-light text-earth-sage' : index === 1 ? 'bg-earth-sky-light text-earth-sky' : index === 2 ? 'bg-earth-clay-light text-earth-clay' : 'text-text-tertiary'"
                    :style="{ background: index > 2 ? 'rgba(0,0,0,0.04)' : undefined }">{{ index + 1 }}</span>
                </td>
                <td class="py-2.5 px-5">
                  <div class="font-medium text-text-primary text-sm">{{ stock.name }}</div>
                  <div class="text-[11px] text-text-tertiary">{{ stock.code }}</div>
                </td>
                <td class="py-2.5 px-5 text-right num-font text-text-primary font-medium text-sm">{{ stock.price.toFixed(2) }}</td>
                <td class="py-2.5 px-5 text-right num-font font-semibold text-sm" :class="stock.changePercent >= 0 ? 'text-up' : 'text-down'">
                  {{ stock.changePercent >= 0 ? '+' : '' }}{{ stock.changePercent.toFixed(2) }}%</td>
                <td class="py-2.5 px-5 text-right">
                  <div class="flex items-center justify-end gap-2">
                    <span class="num-font font-semibold text-sm" :class="stock.totalScore >= 60 ? 'text-earth-sage' : 'text-text-tertiary'">{{ stock.totalScore.toFixed(1) }}</span>
                    <div class="w-12 h-1 bg-black/8 rounded-full overflow-hidden">
                      <div class="h-full rounded-full transition-all" :class="stock.totalScore >= 60 ? 'bg-earth-sage' : 'bg-text-tertiary'" :style="{ width: Math.min(stock.totalScore, 100) + '%' }" />
                    </div>
                  </div>
                </td>
                <td class="py-2.5 px-5 text-center">
                  <span v-if="stock.recommendation" class="text-xs font-medium" :class="getRecommendationStyle(stock.recommendation)">{{ stock.recommendation }}</span>
                  <span v-else class="text-xs text-text-tertiary">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Mobile: cards -->
      <div class="md:hidden px-3 pb-3 pt-2 space-y-2 max-h-[60vh] overflow-y-auto">
        <div v-for="(stock, index) in top10" :key="'m-'+stock.code"
          class="flex items-center gap-2 p-3 bg-white/70 rounded-2xl cursor-pointer active:scale-[0.98] transition-all spring-up"
          :style="{ animationDelay: `${index * 0.04}s` }"
          @click="openDetail(stock.code)">
          <span class="inline-flex items-center justify-center w-7 h-7 rounded-xl text-xs font-semibold shrink-0"
            :class="index === 0 ? 'bg-earth-sage-light text-earth-sage' : index === 1 ? 'bg-earth-sky-light text-earth-sky' : index === 2 ? 'bg-earth-clay-light text-earth-clay' : 'text-text-tertiary'"
            :style="{ background: index > 2 ? 'rgba(0,0,0,0.04)' : undefined }">{{ index + 1 }}</span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-text-primary truncate">{{ stock.name }}</span>
              <span v-if="stock.recommendation" class="text-[9px] font-medium px-1.5 py-0.5 rounded-full shrink-0" :class="getRecommendationStyle(stock.recommendation)">{{ stock.recommendation }}</span>
            </div>
            <div class="flex items-center gap-3 mt-0.5">
              <span class="text-xs text-text-tertiary">{{ stock.code }}</span>
              <span class="num-font text-xs text-text-secondary">¥{{ stock.price.toFixed(2) }}</span>
            </div>
          </div>
          <div class="text-right shrink-0">
            <div class="num-font text-sm font-semibold" :class="stock.changePercent >= 0 ? 'text-up' : 'text-down'">
              {{ stock.changePercent >= 0 ? '+' : '' }}{{ stock.changePercent.toFixed(2) }}%
            </div>
            <div class="flex items-center justify-end gap-1 mt-0.5">
              <span class="num-font text-xs" :class="stock.totalScore >= 60 ? 'text-earth-sage' : 'text-text-tertiary'">{{ stock.totalScore.toFixed(1) }}</span>
              <div class="w-10 h-1 bg-black/8 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all" :class="stock.totalScore >= 60 ? 'bg-earth-sage' : 'bg-text-tertiary'" :style="{ width: Math.min(stock.totalScore, 100) + '%' }" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && !scanResult" class="flex-1 flex items-center justify-center text-text-tertiary">
      <div class="text-center">
        <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" class="mx-auto mb-4 opacity-40"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
        <div class="text-sm font-medium">点击上方按钮扫描全A股</div>
        <div class="text-xs mt-1.5 opacity-70">基于技术面的综合评分与排名</div>
      </div>
    </div>

    <StockDetailSheet :open="detailOpen" :code="selectedCode" @update:open="detailOpen = $event" />
  </div>
</template>
