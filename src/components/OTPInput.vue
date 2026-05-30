<template>
  <div class="flex items-center gap-2 justify-center">
    <input
      v-for="(char, index) in chars"
      :key="index"
      :ref="el => { if (el) inputRefs[index] = el as HTMLInputElement }"
      type="text"
      inputmode="numeric"
      maxlength="1"
      :value="char"
      class="otp-input"
      :class="{ 'otp-input-active': index === activeIndex, 'otp-input-filled': char !== '' }"
      @input="onInput(index, ($event.target as HTMLInputElement).value)"
      @keydown="onKeydown(index, $event)"
      @focus="activeIndex = index"
      @paste="onPaste"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

const props = defineProps<{
  modelValue: string
  length?: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const length = props.length || 6
const chars = ref<string[]>([])
const inputRefs = ref<HTMLInputElement[]>([])
const activeIndex = ref(0)

// Initialize chars from modelValue
onMounted(() => {
  const val = props.modelValue || ''
  chars.value = Array.from({ length }, (_, i) => val[i] || '')
  for (let i = 0; i < chars.value.length; i++) {
    if (chars.value[i] === '') {
      activeIndex.value = i
      break
    }
  }
})

watch(() => props.modelValue, (val) => {
  const newChars = Array.from({ length }, (_, i) => (val || '')[i] || '')
  if (newChars.join('') !== chars.value.join('')) {
    chars.value = newChars
  }
})

const emitValue = () => {
  emit('update:modelValue', chars.value.join(''))
}

const focusInput = (index: number) => {
  if (index >= 0 && index < length) {
    activeIndex.value = index
    inputRefs.value[index]?.focus()
  }
}

const onInput = (index: number, value: string) => {
  // Only keep last digit
  const digit = value.replace(/\D/g, '').slice(-1)
  chars.value[index] = digit
  emitValue()

  if (digit && index < length - 1) {
    focusInput(index + 1)
  }
}

const onKeydown = (index: number, event: KeyboardEvent) => {
  if (event.key === 'Backspace') {
    if (chars.value[index] === '') {
      // If current is empty, go back
      if (index > 0) {
        chars.value[index - 1] = ''
        emitValue()
        focusInput(index - 1)
      }
    } else {
      chars.value[index] = ''
      emitValue()
    }
    event.preventDefault()
  } else if (event.key === 'ArrowLeft' && index > 0) {
    focusInput(index - 1)
    event.preventDefault()
  } else if (event.key === 'ArrowRight' && index < length - 1) {
    focusInput(index + 1)
    event.preventDefault()
  } else if (event.key === 'Enter') {
    // Trigger parent form submit
    const form = (event.target as HTMLElement).closest('form')
    if (form) {
      const btn = form.querySelector('button[type="submit"]') || form.querySelector('button')
      ;(btn as HTMLElement)?.click()
    }
  }
}

const onPaste = (event: ClipboardEvent) => {
  event.preventDefault()
  const paste = event.clipboardData?.getData('text') || ''
  const digits = paste.replace(/\D/g, '').slice(0, length)
  for (let i = 0; i < digits.length; i++) {
    chars.value[i] = digits[i]
  }
  emitValue()
  const nextIndex = Math.min(digits.length, length - 1)
  focusInput(nextIndex)
}

// Expose focus method
defineExpose({ focus: () => focusInput(0) })
</script>

<style scoped>
.otp-input {
  width: 48px;
  height: 56px;
  text-align: center;
  font-size: 24px;
  font-weight: 600;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0;
  background: rgba(39, 39, 42, 0.5);
  border: 1px solid rgba(63, 63, 70, 0.5);
  border-radius: 12px;
  color: #fff;
  outline: none;
  caret-color: #fff;
  transition: all 0.2s ease;
}

.otp-input-active {
  border-color: #fff;
  box-shadow: 0 0 0 1px #fff;
}

.otp-input-filled {
  border-color: rgba(34, 211, 238, 0.4);
  background: rgba(34, 211, 238, 0.05);
}

.otp-input::selection {
  background: rgba(255, 255, 255, 0.2);
}
</style>
