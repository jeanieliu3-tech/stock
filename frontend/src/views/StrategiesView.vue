<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  getStrategyList,
  getStrategy,
  createStrategy,
  updateStrategy,
  deleteStrategy,
  setDefaultStrategy,
  copyStrategy,
  getIndicatorLib,
  evaluateStrategy,
} from '@/api'
import type {
  Strategy,
  StrategyIndicator,
  IndicatorDefinition,
  StockEvalResult,
} from '@/api'

// ─── State ───────────────────────────────
const strategies = ref<Strategy[]>([])
const indicatorLib = ref<IndicatorDefinition[]>([])
const loading = ref(false)
const saving = ref(false)
const previewing = ref(false)
const previewResults = ref<StockEvalResult[]>([])
const error = ref('')

// Editor state
const editingId = ref<string | null>(null)
const creatingNew = ref(false)
const formName = ref('')
const formDesc = ref('')
const formIndicators = ref<FormIndicator[]>([])

interface FormIndicator {
  key: string
  indicatorId: string
  direction: 'up' | 'down'
  weight: number
  minVal: string
  maxVal: string
}

const MAX_INDICATORS = 6

// ─── Computed ────────────────────────────
const totalWeight = computed(() =>
  formIndicators.value.reduce((s, i) => s + i.weight, 0)
)
const weightValid = computed(() => totalWeight.value === 100)
const canAddIndicator = computed(() => formIndicators.value.length < MAX_INDICATORS)
const isEditing = computed(() => editingId.value !== null)

const categoryColors: Record<string, string> = {
  trend: '#E3EDF5',
  oscillator: '#F5EDDF',
  volume: '#E3F0EA',
  volatility: '#F5E3E0',
  pattern: '#EDE6F2',
}

const categoryNames: Record<string, string> = {
  trend: '趋势跟踪',
  oscillator: '动量摆动',
  volume: '量能资金',
  volatility: '波动通道',
  pattern: '支撑压力',
}

const groupedIndicators = computed(() => {
  const groups: { category: string; name: string; items: typeof indicatorLib.value }[] = []
  const catOrder = ['trend', 'oscillator', 'volume', 'volatility', 'pattern']
  for (const cat of catOrder) {
    const items = indicatorLib.value.filter(i => i.category === cat)
    if (items.length > 0) {
      groups.push({ category: cat, name: categoryNames[cat] || cat, items })
    }
  }
  return groups
})

// ─── Lifecycle ───────────────────────────
onMounted(async () => {
  await Promise.all([loadStrategies(), loadIndicatorLib()])
})

// ─── Data loading ────────────────────────
async function loadStrategies() {
  try {
    const res = await getStrategyList()
    if (res.code === 200) strategies.value = res.data || []
  } catch (e) {
    console.error('Failed to load strategies', e)
  }
}

async function loadIndicatorLib() {
  try {
    const res = await getIndicatorLib()
    if (res.code === 200) indicatorLib.value = res.data || []
  } catch (e) {
    console.error('Failed to load indicator lib', e)
  }
}

// ─── Editor actions ──────────────────────
function startNew() {
  editingId.value = null
  creatingNew.value = true
  formName.value = ''
  formDesc.value = ''
  formIndicators.value = []
  previewResults.value = []
  error.value = ''
}

async function startEdit(id: string) {
  try {
    const res = await getStrategy(id)
    if (res.code !== 200 || !res.data) {
      error.value = '获取策略失败'
      return
    }
    const s = res.data
    editingId.value = id
    formName.value = s.name
    formDesc.value = s.description || ''
    formIndicators.value = (s.indicators || []).map((ind) => ({
      key: crypto.randomUUID ? crypto.randomUUID() : `${Date.now()}-${Math.random()}`,
      indicatorId: ind.indicatorId,
      direction: ind.direction,
      weight: ind.weight,
      minVal: ind.minVal || '',
      maxVal: ind.maxVal || '',
    }))
    previewResults.value = []
    error.value = ''
  } catch (e) {
    error.value = '加载策略失败'
  }
}

function cancelEdit() {
  editingId.value = null
  creatingNew.value = false
  formName.value = ''
  formDesc.value = ''
  formIndicators.value = []
  previewResults.value = []
  error.value = ''
}

function addIndicator() {
  if (!canAddIndicator.value) return
  formIndicators.value.push({
    key: crypto.randomUUID ? crypto.randomUUID() : `${Date.now()}-${Math.random()}`,
    indicatorId: indicatorLib.value[0]?.id || '',
    direction: 'up',
    weight: 0,
    minVal: '',
    maxVal: '',
  })
}

function removeIndicator(key: string) {
  formIndicators.value = formIndicators.value.filter((i) => i.key !== key)
}

function getIndicatorName(id: string): string {
  const ind = indicatorLib.value.find((i) => i.id === id)
  return ind?.name || id
}

function getIndicatorCategory(id: string): string {
  const ind = indicatorLib.value.find((i) => i.id === id)
  return ind?.category || ''
}

function getIndicatorCardBg(id: string): string {
  const cat = getIndicatorCategory(id)
  return categoryColors[cat] || 'rgba(0,0,0,0.04)'
}

const catTextColors: Record<string, string> = {
  trend: '#4A7A9E',
  oscillator: '#A8884A',
  volume: '#4A8A6A',
  volatility: '#A86B62',
  pattern: '#7A62A0',
}

function getCatTextColor(id: string): string {
  const cat = getIndicatorCategory(id)
  return catTextColors[cat] || '#6B6B6B'
}

async function handleSetDefault(id: string) {
  try {
    const res = await setDefaultStrategy(id)
    if (res.code === 200) await loadStrategies()
  } catch (e) {
    console.error('Set default failed', e)
  }
}

async function handleCopy(id: string) {
  try {
    const res = await copyStrategy(id)
    if (res.code === 200) await loadStrategies()
  } catch (e) {
    console.error('Copy failed', e)
  }
}

async function handleDelete(id: string) {
  if (!confirm('确定删除此策略？')) return
  try {
    const res = await deleteStrategy(id)
    if (res.code === 200) {
      await loadStrategies()
      if (editingId.value === id) cancelEdit()
    }
  } catch (e) {
    console.error('Delete failed', e)
  }
}

async function handleSave() {
  if (!formName.value.trim()) {
    error.value = '请输入策略名称'
    return
  }
  if (formIndicators.value.length === 0) {
    error.value = '请至少添加一个指标'
    return
  }
  if (!weightValid.value) {
    error.value = '权重之和必须等于 100%'
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload = {
      name: formName.value.trim(),
      description: formDesc.value.trim(),
      indicators: formIndicators.value.map((ind) => ({
        indicatorId: ind.indicatorId,
        direction: ind.direction,
        weight: ind.weight,
        minVal: ind.minVal,
        maxVal: ind.maxVal,
      })),
    }
    if (editingId.value) {
      const res = await updateStrategy(editingId.value, payload)
      if (res.code !== 200) {
        error.value = res.msg || '更新失败'
        return
      }
    } else {
      const res = await createStrategy(payload)
      if (res.code !== 200) {
        error.value = res.msg || '创建失败'
        return
      }
    }
    await loadStrategies()
    cancelEdit()
  } catch (e) {
    error.value = '保存失败'
  } finally {
    saving.value = false
  }
}

async function handlePreview() {
  if (formIndicators.value.length === 0) return
  previewing.value = true
  error.value = ''
  try {
    const payload = {
      strategy: {
        name: formName.value.trim() || '预览',
        description: formDesc.value.trim(),
        indicators: formIndicators.value.map((ind) => ({
          indicatorId: ind.indicatorId,
          direction: ind.direction,
          weight: ind.weight,
          minVal: ind.minVal,
          maxVal: ind.maxVal,
        })),
      },
      page: 1,
      pageSize: 5,
    }
    const res = await evaluateStrategy(payload)
    if (res.code === 200 && res.data) {
      previewResults.value = res.data.items || []
    } else {
      error.value = res.msg || '评估失败'
    }
  } catch (e) {
    error.value = '评估请求失败'
  } finally {
    previewing.value = false
  }
}

function truncate(s: string, n: number): string {
  return s && s.length > n ? s.slice(0, n) + '...' : s
}

function indicatorCount(s: Strategy): number {
  return s.indicators?.length || 0
}
</script>

<template>
  <div style="max-width: 1100px; margin: 0 auto;">
    <div class="flex flex-col md:flex-row gap-5 md:gap-8" style="min-height:calc(100vh - 8rem)">
    <!-- ═══ Left: Strategy List ═══ -->
    <div class="w-full md:w-[300px] shrink-0 flex flex-col gap-5">
      <div class="flex items-center justify-between">
        <h2 class="text-base font-semibold text-text-primary">策略库</h2>
        <button
          class="flex items-center gap-1.5 px-4 py-2 rounded-2xl text-sm font-medium btn-primary transition-all duration-300"
          @click="startNew"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新建
        </button>
      </div>

      <!-- Mobile: show list when not editing -->
      <div class="md:block" :class="editingId !== null || creatingNew ? 'hidden' : 'block'">
        <div class="flex-1 overflow-y-auto space-y-2.5 pr-1" style="max-height:calc(100vh - 10rem)">
          <div
            v-for="s in strategies"
            :key="s.id"
            class="rounded-2xl p-4 cursor-pointer transition-all duration-300 hover-lift card"
            :class="editingId === s.id ? 'card' : ''"
            @click="startEdit(s.id)"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-semibold text-text-primary">{{ s.name }}</span>
                  <svg v-if="s.isDefault" width="14" height="14" viewBox="0 0 24 24" fill="#C4A88C" stroke="#C4A88C" stroke-width="1.5"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                </div>
                <div v-if="s.description" class="text-xs mt-1 text-text-tertiary leading-relaxed">{{ truncate(s.description, 36) }}</div>
              </div>
              <span class="shrink-0 text-[11px] font-medium px-2.5 py-0.5 rounded-full bg-earth-sky-light text-earth-sky">{{ indicatorCount(s) }}</span>
            </div>

            <div v-if="!s.isSystem" class="flex items-center gap-2 mt-3 pt-3" style="border-top:1px solid rgba(0,0,0,0.04);">
              <button class="flex items-center gap-1 text-[11px] px-2.5 py-1 rounded-xl transition-colors text-text-tertiary hover:text-text-secondary hover:bg-black/5" @click.stop="handleSetDefault(s.id)">设为默认</button>
              <button class="flex items-center gap-1 text-[11px] px-2.5 py-1 rounded-xl transition-colors text-text-tertiary hover:text-text-secondary hover:bg-black/5" @click.stop="handleCopy(s.id)">复制</button>
              <button class="flex items-center gap-1 text-[11px] px-2.5 py-1 rounded-xl transition-colors ml-auto text-earth-terracotta hover:bg-earth-terracotta-light/50" @click.stop="handleDelete(s.id)">删除</button>
            </div>
          </div>

          <div v-if="strategies.length === 0" class="text-center py-12 text-sm text-text-tertiary">
            暂无策略，点击新建创建
          </div>
        </div>
      </div>
    </div>

    <!-- ═══ Right: Editor ═══ -->
    <div class="flex-1 min-w-0">
      <!-- Mobile: back button when editing -->
      <div class="md:hidden flex items-center gap-3 mb-4">
        <button class="flex items-center gap-1.5 px-3 py-2 rounded-2xl bg-black/5 text-sm text-text-secondary hover:bg-black/10 transition-all"
          @click="cancelEdit">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="15 18 9 12 15 6"/></svg>
          返回策略列表
        </button>
      </div>

      <!-- Empty state -->
      <div v-if="editingId === null && !creatingNew" class="h-full flex items-center justify-center">
        <div class="text-center" style="max-width:280px">
          <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" class="mx-auto mb-4 text-text-tertiary/40"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          <p class="text-text-tertiary text-sm leading-relaxed">选择一个策略编辑，<br>或点击左侧「新建」</p>
        </div>
      </div>

      <!-- Editor -->
      <div v-else class="space-y-7 w-full">
        <!-- Error -->
        <Transition name="page">
          <div v-if="error" class="flex items-center gap-2.5 px-5 py-3.5 rounded-2xl bg-[#F5E3E0]">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#A86B62" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
            <span class="text-sm text-[#A86B62]">{{ error }}</span>
          </div>
        </Transition>

        <!-- Basic Info -->
        <div class="card p-5 md:p-7 space-y-5">
          <h3 class="text-sm font-semibold text-text-primary">基本信息</h3>
          <div class="space-y-4">
            <div>
              <label class="text-xs font-medium mb-1.5 block text-text-secondary">策略名称</label>
              <input v-model="formName" placeholder="输入策略名称" class="input-organic w-full text-sm" />
            </div>
            <div>
              <label class="text-xs font-medium mb-1.5 block text-text-secondary">描述</label>
              <textarea v-model="formDesc" placeholder="简要描述策略用途（可选）" rows="2"
                class="input-organic w-full text-sm resize-none" style="min-height:60px" />
            </div>
          </div>
        </div>

        <!-- Indicators -->
        <div class="card p-5 md:p-7 space-y-5">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-semibold text-text-primary">指标配置</h3>
            <span class="text-xs text-text-tertiary">最多 {{ MAX_INDICATORS }} 个</span>
          </div>

          <div class="space-y-3">
            <div
              v-for="(ind, idx) in formIndicators"
              :key="ind.key"
              class="rounded-2xl p-4 md:p-6 transition-all duration-300"
              :style="{
                background: `linear-gradient(135deg, ${getIndicatorCardBg(ind.indicatorId)} 60%, #fff 100%)`,
                border: '1px solid rgba(0,0,0,0.04)',
              }"
            >
              <div class="flex items-center justify-between mb-4">
                <span class="text-xs font-semibold" :style="{ color: getCatTextColor(ind.indicatorId) }">指标 {{ idx + 1 }}</span>
                <button
                  class="flex items-center gap-1 text-xs px-2.5 py-1 rounded-xl transition-all text-text-tertiary hover:text-text-secondary"
                  style="background:rgba(255,255,255,0.6)"
                  @click="removeIndicator(ind.key)"
                >
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  移除
                </button>
              </div>

              <div class="flex flex-wrap items-start gap-4">
                <div class="flex-1" style="min-width:180px">
                  <label class="text-[11px] font-medium mb-1.5 block" style="color:rgba(0,0,0,0.45)">选择指标</label>
                  <select v-model="ind.indicatorId"
                    class="w-full px-3.5 py-2 rounded-2xl text-sm outline-none transition-all"
                    style="border:1px solid rgba(0,0,0,0.08);color:#2C2C2C;background:rgba(255,255,255,0.8);height:38px"
                  >
                    <option value="" disabled>请选择</option>
                    <optgroup v-for="group in groupedIndicators" :key="group.category" :label="group.name">
                      <option v-for="lib in group.items" :key="lib.id" :value="lib.id">{{ lib.name }}</option>
                    </optgroup>
                  </select>
                </div>

                <div>
                  <label class="text-[11px] font-medium mb-1.5 block" style="color:rgba(0,0,0,0.45)">方向</label>
                  <div class="flex rounded-2xl overflow-hidden" style="border:1px solid rgba(0,0,0,0.08);height:32px;background:rgba(255,255,255,0.5)">
                    <button
                      class="px-3.5 text-xs font-medium transition-all cursor-pointer" style="line-height:32px;border:none;background:transparent"
                      :style="ind.direction === 'up' ? { background:'rgba(141,184,160,0.15)', color:'#8DB8A0', fontWeight:600 } : { color:'rgba(0,0,0,0.35)' }"
                      @click="ind.direction = 'up'"
                    >↑ 越大</button>
                    <button
                      class="px-3.5 text-xs font-medium transition-all cursor-pointer" style="line-height:32px;border:none;background:transparent;border-left:1px solid rgba(0,0,0,0.06)"
                      :style="ind.direction === 'down' ? { background:'rgba(141,184,160,0.15)', color:'#8DB8A0', fontWeight:600 } : { color:'rgba(0,0,0,0.35)' }"
                      @click="ind.direction = 'down'"
                    >↓ 越小</button>
                  </div>
                </div>

                <div>
                  <label class="text-[11px] font-medium mb-1.5 block" style="color:rgba(0,0,0,0.45)">权重</label>
                  <div class="flex items-center gap-2" style="height:32px">
                    <input type="range" min="0" max="100" step="1" v-model.number="ind.weight"
                      class="cursor-pointer w-full md:w-[80px]" style="height:4px;border-radius:2px;appearance:none;background:rgba(0,0,0,0.1);outline:none;accent-color:#8DB8A0"
                    />
                    <span class="num-font text-xs font-semibold text-text-primary" style="width:28px;text-align:right">{{ ind.weight }}%</span>
                  </div>
                </div>
              </div>

              <div class="flex gap-4 mt-3 pt-3" style="border-top:1px solid rgba(0,0,0,0.04);">
                <div class="flex-1 flex items-center gap-2">
                  <span class="text-[11px] shrink-0" style="color:rgba(0,0,0,0.35)">≥</span>
                  <input v-model="ind.minVal" placeholder="不限"
                    class="w-full px-3 py-1.5 rounded-xl text-xs outline-none num-font transition-all"
                    style="height:32px;border:1px solid rgba(0,0,0,0.07);color:#2C2C2C;background:rgba(255,255,255,0.6)"
                  />
                </div>
                <div class="flex-1 flex items-center gap-2">
                  <span class="text-[11px] shrink-0" style="color:rgba(0,0,0,0.35)">≤</span>
                  <input v-model="ind.maxVal" placeholder="不限"
                    class="w-full px-3 py-1.5 rounded-xl text-xs outline-none num-font transition-all"
                    style="height:32px;border:1px solid rgba(0,0,0,0.07);color:#2C2C2C;background:rgba(255,255,255,0.6)"
                  />
                </div>
              </div>
            </div>

            <div v-if="formIndicators.length === 0" class="text-center py-8 text-sm text-text-tertiary">
              尚未添加指标，点击下方按钮添加
            </div>
          </div>

          <div class="flex items-center justify-between pt-4" style="border-top:1px solid rgba(0,0,0,0.04);">
            <div class="flex items-center gap-2">
              <span class="text-xs text-text-secondary">权重合计</span>
              <span class="num-font text-sm font-semibold" :class="weightValid ? 'text-down' : 'text-up'">{{ totalWeight }}%</span>
              <span v-if="!weightValid" class="text-[11px] text-up">须等于 100%</span>
            </div>
            <button
              class="text-xs font-medium px-4 py-1.5 rounded-full transition-all duration-300 cursor-pointer"
              :class="canAddIndicator ? 'text-earth-sage border-earth-sage bg-white' : 'text-text-tertiary border-black/10 bg-black/4'"
              :style="{ border: '1.5px solid' }"
              :disabled="!canAddIndicator"
              @click="addIndicator"
            >+ 添加指标</button>
          </div>
        </div>

        <!-- Preview -->
        <div class="card p-5 md:p-7 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-semibold text-text-primary">评分预览</h3>
            <button
              class="flex items-center gap-1.5 text-xs font-medium px-4 py-2 rounded-2xl transition-all duration-300 btn-primary"
              :class="{ 'opacity-50': formIndicators.length === 0 || previewing }"
              :disabled="formIndicators.length === 0 || previewing"
              @click="handlePreview"
            >
              <svg v-if="previewing" class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
              <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              {{ previewing ? '计算中...' : '评分预览' }}
            </button>
          </div>

          <div v-if="previewResults.length > 0" class="space-y-2">
            <div v-for="(r, idx) in previewResults" :key="r.code"
              class="flex items-center justify-between px-4 py-3 rounded-2xl transition-all bg-black/4"
            >
              <div class="flex items-center gap-3">
                <span class="num-font text-xs font-semibold text-text-tertiary" style="width:20px">{{ idx + 1 }}</span>
                <div>
                  <span class="text-sm font-semibold text-text-primary">{{ r.name }}</span>
                  <span class="text-xs ml-2 text-text-tertiary">{{ r.code }}</span>
                </div>
              </div>
              <div class="flex items-center gap-4">
                <span class="num-font text-xs text-text-secondary">¥{{ r.price.toFixed(2) }}</span>
                <span class="num-font text-xs font-semibold" :class="r.changePercent >= 0 ? 'text-up' : 'text-down'">
                  {{ r.changePercent >= 0 ? '+' : '' }}{{ r.changePercent.toFixed(2) }}%
                </span>
                <span class="num-font text-sm font-semibold" :class="r.totalScore >= 70 ? 'text-down' : r.totalScore >= 40 ? 'text-earth-clay' : 'text-up'">
                  {{ r.totalScore.toFixed(1) }}
                </span>
              </div>
            </div>
          </div>
          <div v-else class="text-center py-8 text-sm text-text-tertiary">
            点击「评分预览」查看策略评分 Top 5
          </div>
        </div>

        <!-- Actions -->
        <div class="flex items-center justify-end gap-3 pb-8">
          <button class="px-5 py-2.5 rounded-2xl text-sm font-medium transition-all bg-black/5 text-text-secondary hover:bg-black/10" @click="cancelEdit">取消</button>
          <button
            class="px-5 py-2.5 rounded-2xl text-sm font-medium transition-all flex items-center gap-2 btn-primary"
            :class="{ 'opacity-50': saving }"
            :disabled="saving"
            @click="handleSave"
          >
            <svg v-if="saving" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            {{ saving ? '保存中...' : '保存策略' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
</template>
