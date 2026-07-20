<script setup lang="ts">
import { computed } from 'vue'
import { getTagStyle, getRecommendationStyle, getMacdSignalStyle } from '@/utils/tagColors'

const props = withDefaults(defineProps<{
  type?: 'highlight' | 'recommendation' | 'macdSignal' | 'bollPosition' | 'signalType' | 'score'
  text: string
  size?: 'sm' | 'md'
  score?: number
}>(), {
  type: 'highlight',
  size: 'sm',
  score: 0,
})

const styleClass = computed(() => {
  switch (props.type) {
    case 'recommendation': {
      return getRecommendationStyle(props.text)
    }
    case 'macdSignal': {
      const s = getMacdSignalStyle(props.text)
      return `${s.bg} ${s.text}`
    }
    case 'bollPosition': {
      return BOLL_MAP[props.text] || 'text-text-tertiary'
    }
    case 'signalType': {
      return SIG_MAP[props.text] || 'text-text-tertiary'
    }
    case 'score': {
      if (props.score >= 60) return 'text-earth-sage font-semibold'
      return 'text-text-tertiary'
    }
    default: {
      const s = getTagStyle(props.text)
      return `${s.bg} ${s.text}`
    }
  }
})

const sizeClass = computed(() => {
  return props.size === 'md' ? 'text-sm px-3 py-1.5' : 'text-xs px-2.5 py-1'
})

const useTag = computed(() => {
  return props.type === 'recommendation' ? false : true
})

const BOLL_MAP: Record<string, string> = {
  '突破上轨': 'text-[#A86B62] font-medium',
  '上轨区域': 'text-[#A8884A]',
  '中轨上方': 'text-[#4A7A9E]',
  '中轨下方': 'text-[#6B6B6B]',
  '下轨区域': 'text-[#6B6B6B]',
  '跌破下轨': 'text-[#4A8A6A]',
  '待分析':   'text-[#9A9A9A]',
}
const SIG_MAP: Record<string, string> = {
  'sell':    'text-[#A86B62]',
  'buy':     'text-[#4A8A6A]',
  'hold':    'text-[#4A7A9E]',
  'warning': 'text-[#A8884A]',
}
</script>

<template>
  <span v-if="useTag" :class="['tag-organic', sizeClass, styleClass]">
    <slot>{{ text }}</slot>
  </span>
  <span v-else :class="['text-xs font-medium', sizeClass, styleClass]">
    <slot>{{ text }}</slot>
  </span>
</template>
