<template>
  <div class="space-y-5">
    <h1 class="text-[28px] font-semibold text-text-primary tracking-tight">全A排名</h1>

    <!-- 统计信息 -->
    <div v-if="scanResult" class="flex items-center gap-4 text-xs text-text-tertiary bg-white/50 rounded-2xl px-5 py-3 flex-wrap">
      <span>总数 <span class="num-font font-semibold text-text-secondary">{{ scanResult.totalStocks.toLocaleString() }}</span></span>
      <span>有效 <span class="num-font font-semibold text-text-secondary">{{ scanResult.validStocks.toLocaleString() }}</span></span>
      <span>耗时 <span class="num-font font-semibold text-text-secondary">{{ (scanResult.costMs / 1000).toFixed(1) }}s</span></span>
      <span class="text-text-tertiary">{{ scanResult.scanTime }}</span>
    </div>

    <!-- 操作栏 -->
    <div class="glass rounded-2xl p-4 flex flex-col sm:flex-row items-start sm:items-center gap-3">
      <button class="btn btn-primary text-sm px-5 py-2.5 flex items-center gap-2 disabled:opacity-50" :disabled="scanning" @click="startScan">
        <svg v-if="scanning" class="animate-spin" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10" stroke-dasharray="30 70" stroke-linecap="round"/></svg>
        <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
        {{ scanning ? '扫描中...' : '全A股扫描' }}
      </button>
      <div class="flex-1" />
      <div class="flex items-center gap-2 flex-wrap">
        <select v-model="sortBy" class="text-xs rounded-2xl px-3.5 py-2 bg-white/70 outline-none text-text-secondary transition-all" style="border:1px solid rgba(0,0,0,0.06);font-family:'Outfit','PingFang SC',sans-serif;">
          <option value="totalScore">综合评分</option>
          <option value="changePercent">涨跌幅</option>
          <option value="turnoverRate">换手率</option>
          <option value="volumeRatio">量比</option>
          <option value="techScore">技术评分</option>
          <option value="momentumScore">动量评分</option>
        </select>
        <select v-model="filterType" class="text-xs rounded-2xl px-3.5 py-2 bg-white/70 outline-none text-text-secondary transition-all" style="border:1px solid rgba(0,0,0,0.06);font-family:'Outfit','PingFang SC',sans-serif;">
          <option value="">全部信号</option>
          <option value="golden_cross">MACD金叉</option>
          <option value="above_water">水上金叉</option>
          <option value="strong">强势股</option>
          <option value="volume_break">放量突破</option>
        </select>
        <select v-model="scoreFilter" class="text-xs rounded-2xl px-3.5 py-2 bg-white/70 outline-none text-text-secondary transition-all" style="border:1px solid rgba(0,0,0,0.06);font-family:'Outfit','PingFang SC',sans-serif;">
          <option value="">评分阈值</option>
          <option value="75">≥75</option>
          <option value="60">≥60</option>
          <option value="45">≥45</option>
          <option value="30">≥30</option>
        </select>
        <select v-model="pageSize" class="text-xs rounded-2xl px-3.5 py-2 bg-white/70 outline-none text-text-secondary transition-all" style="border:1px solid rgba(0,0,0,0.06);font-family:'Outfit','PingFang SC',sans-serif;">
          <option :value="20">20条</option>
          <option :value="50">50条</option>
          <option :value="100">100条</option>
        </select>
      </div>
    </div>

    <!-- 错误 -->
    <div v-if="errorMsg" class="bg-[#F5E3E0] rounded-2xl px-5 py-3.5 text-sm text-[#A86B62] flex items-center justify-between">
      <span>{{ errorMsg }}</span>
      <button class="underline text-xs font-medium" @click="errorMsg = ''">关闭</button>
    </div>

    <!-- 排名表格（桌面） / 卡片列表（移动端） -->
    <div class="card overflow-hidden" style="max-height: calc(100vh - 280px);">

      <!-- Desktop: table -->
      <div class="hidden md:block overflow-y-auto" style="max-height: calc(100vh - 280px);">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr>
                <th class="text-center text-xs font-medium text-text-tertiary py-3.5 px-4 w-9 sticky top-0 bg-white/90 backdrop-blur-md z-10">#</th>
                <th class="text-left text-xs font-medium text-text-tertiary py-3.5 px-4 min-w-[110px] sticky top-0 bg-white/90 backdrop-blur-md z-10">股票</th>
                <th class="text-right text-xs font-medium text-text-tertiary py-3.5 px-4 num-font sticky top-0 bg-white/90 backdrop-blur-md z-10">现价</th>
                <th class="text-right text-xs font-medium text-text-tertiary py-3.5 px-4 num-font sticky top-0 bg-white/90 backdrop-blur-md z-10">涨跌幅</th>
                <th class="text-right text-xs font-medium text-text-tertiary py-3.5 px-4 hidden md:table-cell num-font sticky top-0 bg-white/90 backdrop-blur-md z-10">量比</th>
                <th class="text-right text-xs font-medium text-text-tertiary py-3.5 px-4 hidden lg:table-cell num-font sticky top-0 bg-white/90 backdrop-blur-md z-10">换手率</th>
                <th class="text-right text-xs font-medium text-text-tertiary py-3.5 px-4 sticky top-0 bg-white/90 backdrop-blur-md z-10">综合评分</th>
                <th class="text-center text-xs font-medium text-text-tertiary py-3.5 px-4 hidden sm:table-cell sticky top-0 bg-white/90 backdrop-blur-md z-10">亮点</th>
                <th class="text-center text-xs font-medium text-text-tertiary py-3.5 px-4 sticky top-0 bg-white/90 backdrop-blur-md z-10">建议</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in rankData?.items ?? []" :key="item.code"
                class="cursor-pointer transition-all duration-200 hover:bg-black/4 spring-up"
                :style="{ animationDelay: `${(item.rank - 1) * 0.025}s` }"
                @click="openDetail(item.code)">
              <td class="py-3 px-4 text-center">
                <span class="inline-flex items-center justify-center w-7 h-7 rounded-xl text-xs font-semibold"
                  :class="item.rank === 1 ? 'bg-earth-sage-light text-earth-sage' : item.rank === 2 ? 'bg-earth-sky-light text-earth-sky' : item.rank === 3 ? 'bg-earth-clay-light text-earth-clay' : 'text-text-tertiary'"
                  :style="{ background: item.rank > 3 ? 'rgba(0,0,0,0.04)' : undefined }">{{ item.rank }}</span>
              </td>
              <td class="py-3 px-4">
                <div class="flex items-center gap-2">
                  <button class="shrink-0 text-sm leading-none transition-all duration-200" :class="isFavorite(item.code) ? 'text-earth-sage' : 'text-text-tertiary/30 hover:text-text-tertiary'" @click.stop="toggleFavorite(item.code)">
                    <svg width="14" height="14" viewBox="0 0 24 24" :fill="isFavorite(item.code) ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                  </button>
                  <div><div class="font-medium text-text-primary text-sm">{{ item.name }}</div><div class="text-xs text-text-tertiary">{{ item.code }}</div></div>
                </div>
              </td>
              <td class="py-3 px-4 text-right num-font text-text-primary font-medium">{{ item.price.toFixed(2) }}</td>
              <td class="py-3 px-4 text-right num-font font-semibold" :class="item.changePercent > 0 ? 'text-up' : item.changePercent < 0 ? 'text-down' : 'text-text-tertiary'">{{ item.changePercent > 0 ? '+' : '' }}{{ item.changePercent.toFixed(2) }}%</td>
              <td class="py-3 px-4 text-right num-font hidden md:table-cell">
                <span v-if="item.volumeRatio > 0" :class="item.volumeRatio >= 3 ? 'text-up font-medium' : item.volumeRatio >= 2 ? 'text-[#A8884A]' : 'text-text-secondary'">{{ item.volumeRatio.toFixed(2) }}</span>
                <span v-else class="text-text-tertiary">-</span>
              </td>
              <td class="py-3 px-4 text-right num-font hidden lg:table-cell">
                <span v-if="item.turnoverRate > 0" class="text-text-secondary">{{ item.turnoverRate.toFixed(2) }}%</span>
                <span v-else class="text-text-tertiary">-</span>
              </td>
              <td class="py-3 px-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <span class="num-font font-semibold text-sm" :class="item.totalScore >= 60 ? 'text-earth-sage' : 'text-text-tertiary'">{{ item.totalScore.toFixed(1) }}</span>
                  <div class="w-14 h-1 bg-black/8 rounded-full overflow-hidden shrink-0">
                    <div class="h-full rounded-full transition-all duration-500" :class="item.totalScore >= 60 ? 'bg-earth-sage' : 'bg-text-tertiary'" :style="{ width: Math.min(item.totalScore, 100) + '%' }" />
                  </div>
                </div>
              </td>
              <td class="py-3 px-4 text-center hidden sm:table-cell">
                <div class="flex flex-wrap gap-1 justify-center">
                  <span v-for="h in item.highlights?.slice(0, 2)" :key="h" class="tag-organic" :class="getHighlightStyle(h)">{{ h }}</span>
                </div>
              </td>
              <td class="py-3 px-4 text-center">
                <span v-if="item.recommendation" class="text-xs font-medium" :class="getRecommendationStyle(item.recommendation)">{{ item.recommendation }}</span>
                <span v-else class="text-xs text-text-tertiary">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Mobile: card list -->
      <div class="md:hidden overflow-y-auto px-3 py-2 space-y-2" style="max-height: calc(100vh - 320px);">
        <div v-for="item in rankData?.items ?? []" :key="'m-'+item.code"
          class="flex items-center gap-2 p-3 bg-white/70 rounded-2xl cursor-pointer active:scale-[0.98] transition-all spring-up"
          :style="{ animationDelay: `${(item.rank - 1) * 0.025}s` }"
          @click="openDetail(item.code)">
          <span class="inline-flex items-center justify-center w-7 h-7 rounded-xl text-xs font-semibold shrink-0"
            :class="item.rank === 1 ? 'bg-earth-sage-light text-earth-sage' : item.rank === 2 ? 'bg-earth-sky-light text-earth-sky' : item.rank === 3 ? 'bg-earth-clay-light text-earth-clay' : 'text-text-tertiary'"
            :style="{ background: item.rank > 3 ? 'rgba(0,0,0,0.04)' : undefined }">{{ item.rank }}</span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-text-primary truncate">{{ item.name }}</span>
              <button class="shrink-0" :class="isFavorite(item.code) ? 'text-earth-sage' : 'text-text-tertiary/20'"
                @click.stop="toggleFavorite(item.code)">
                <svg width="11" height="11" viewBox="0 0 24 24" :fill="isFavorite(item.code) ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
              </button>
              <span v-if="item.recommendation" class="text-[9px] font-medium px-1.5 py-0.5 rounded-full shrink-0 ml-auto" :class="getRecommendationStyle(item.recommendation)">{{ item.recommendation }}</span>
            </div>
            <div class="flex items-center gap-3 mt-0.5 text-xs text-text-tertiary">
              <span>{{ item.code }}</span>
              <span class="num-font">¥{{ item.price.toFixed(2) }}</span>
            </div>
          </div>
          <div class="text-right shrink-0">
            <div class="num-font text-sm font-semibold" :class="item.changePercent > 0 ? 'text-up' : item.changePercent < 0 ? 'text-down' : 'text-text-tertiary'">
              {{ item.changePercent > 0 ? '+' : '' }}{{ item.changePercent.toFixed(2) }}%
            </div>
            <div class="flex items-center justify-end gap-1 mt-0.5">
              <span class="num-font text-xs" :class="item.totalScore >= 60 ? 'text-earth-sage' : 'text-text-tertiary'">{{ item.totalScore.toFixed(1) }}</span>
              <div class="w-10 h-1 bg-black/8 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all" :class="item.totalScore >= 60 ? 'bg-earth-sage' : 'bg-text-tertiary'" :style="{ width: Math.min(item.totalScore, 100) + '%' }" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 空/加载状态 -->
      <div v-if="!rankData?.items?.length && !scanning" class="py-20 text-center text-text-tertiary">
        <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" class="mx-auto mb-4 opacity-40"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
        <p class="text-sm">点击「全A股扫描」开始分析</p>
        <p class="text-xs mt-1 opacity-70">扫描约5000支A股，评分排名约需30-60秒</p>
      </div>

      <div v-if="scanning" class="py-20 text-center">
        <svg class="animate-spin mx-auto mb-3 text-earth-sage" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><circle cx="12" cy="12" r="10" stroke-dasharray="30 70"/></svg>
        <p class="text-sm text-text-secondary">正在扫描全A股...</p>
      </div>
    </div>
    </div>

    <!-- 分页（桌面） -->
    <div v-if="rankData && rankData.totalPages > 1" class="hidden md:flex items-center justify-between">
      <div class="text-xs text-text-tertiary">共 <span class="font-medium text-text-secondary">{{ rankData.total }}</span> 支，第 {{ rankData.page }}/{{ rankData.totalPages }} 页</div>
      <div class="flex items-center gap-2">
        <button class="btn-primary text-xs px-3.5 py-2 rounded-2xl disabled:opacity-40" :disabled="currentPage <= 1" @click="currentPage = 1">首页</button>
        <button class="bg-white/70 text-xs px-3.5 py-2 rounded-2xl disabled:opacity-40 transition-all hover:bg-white/90" style="border:1px solid rgba(0,0,0,0.06);" :disabled="currentPage <= 1" @click="currentPage--">上一页</button>
        <span class="text-xs text-text-tertiary px-2">{{ currentPage }}/{{ rankData.totalPages }}</span>
        <button class="bg-white/70 text-xs px-3.5 py-2 rounded-2xl disabled:opacity-40 transition-all hover:bg-white/90" style="border:1px solid rgba(0,0,0,0.06);" :disabled="currentPage >= rankData.totalPages" @click="currentPage++">下一页</button>
        <button class="btn-primary text-xs px-3.5 py-2 rounded-2xl disabled:opacity-40" :disabled="currentPage >= rankData.totalPages" @click="currentPage = rankData.totalPages">末页</button>
      </div>
    </div>

    <!-- 分页（移动端）- 简化 -->
    <div v-if="rankData && rankData.totalPages > 1" class="md:hidden flex items-center justify-between bg-white/60 rounded-2xl px-4 py-3">
      <div class="text-xs text-text-tertiary">共 {{ rankData.total }} 支</div>
      <div class="flex items-center gap-2">
        <button class="w-9 h-9 flex items-center justify-center rounded-2xl bg-white/70 text-text-secondary disabled:opacity-30 transition-all" style="border:1px solid rgba(0,0,0,0.06);" :disabled="currentPage <= 1" @click="currentPage--">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="15 18 9 12 15 6"/></svg>
        </button>
        <span class="text-xs text-text-secondary px-2">{{ currentPage }}/{{ rankData.totalPages }}</span>
        <button class="w-9 h-9 flex items-center justify-center rounded-2xl bg-white/70 text-text-secondary disabled:opacity-30 transition-all" style="border:1px solid rgba(0,0,0,0.06);" :disabled="currentPage >= rankData.totalPages" @click="currentPage++">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="9 18 15 12 9 6"/></svg>
        </button>
      </div>
    </div>

    <StockDetailSheet :open="!!detailCode" :code="detailCode" @update:open="detailCode = $event ? detailCode : ''" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { scanAllAShares, getAllStockRank } from '@/api'
import type { AllStockScanResult, AllStockRankResponse } from '@/types/stock'
import StockDetailSheet from '@/components/StockDetailSheet.vue'
import { getRecommendationStyle, getTagStyle } from '@/utils/tagColors'

const scanning = ref(false)
const scanResult = ref<AllStockScanResult | null>(null)
const rankData = ref<AllStockRankResponse | null>(null)
const currentPage = ref(1)
const pageSize = ref(20)
const sortBy = ref('totalScore')
const filterType = ref('')
const changeFilter = ref('')
const scoreFilter = ref('')
const detailCode = ref('')
const errorMsg = ref('')

const favorites = ref<string[]>(JSON.parse(localStorage.getItem('stock_favorites') || '[]'))

function isFavorite(code: string) { return favorites.value.includes(code) }
function toggleFavorite(code: string) {
  if (isFavorite(code)) favorites.value = favorites.value.filter(c => c !== code)
  else favorites.value.push(code)
  localStorage.setItem('stock_favorites', JSON.stringify(favorites.value))
}

function getHighlightStyle(tag: string): string {
  const s = getTagStyle(tag)
  return `${s.bg} ${s.text}`
}

const openDetail = (code: string) => { detailCode.value = code }

async function startScan() {
  scanning.value = true; errorMsg.value = ''
  try {
    const res = await scanAllAShares()
    if (res.code === 200 && res.data) { scanResult.value = res.data; currentPage.value = 1; await loadRank() }
    else { errorMsg.value = '扫描返回异常：' + (res.msg || '未知错误') }
  } catch (e: any) { errorMsg.value = '全A扫描失败：' + (e?.message || '网络超时或服务异常') }
  finally { scanning.value = false }
}

async function loadRank() {
  try {
    const params: Record<string, string | number> = { page: currentPage.value, pageSize: pageSize.value, sortBy: sortBy.value, order: 'desc' }
    let filter = filterType.value
    if (changeFilter.value) { const v = parseFloat(changeFilter.value); filter += (filter ? ',' : '') + (v > 0 ? `change_gt_${v}` : `change_lt_${Math.abs(v)}`) }
    if (scoreFilter.value) filter += (filter ? ',' : '') + `score_gte_${scoreFilter.value}`
    if (filter) (params as any).filter = filter
    const res = await getAllStockRank(params)
    if (res.code === 200 && res.data) rankData.value = res.data as AllStockRankResponse
  } catch (e) { console.error('加载排名失败', e) }
}

watch([currentPage, pageSize, sortBy, filterType, changeFilter, scoreFilter], () => { if (scanResult.value) loadRank() })
onMounted(() => { loadRank() })
</script>
