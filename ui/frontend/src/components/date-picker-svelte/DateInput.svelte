<script>import { fly } from 'svelte/transition';
import { cubicInOut } from 'svelte/easing';
import { toText } from './date-utils';
import { parse, createFormat } from './parse';
import DateTimePicker from './DatePicker.svelte';
import { writable } from 'svelte/store';
import { createEventDispatcher } from 'svelte';
const dispatch = createEventDispatcher();
/** Default date to display in picker before value is assigned */
const defaultDate = new Date();
// inner date value store for preventing value updates (and also
// text updates as a result) when date is unchanged
const innerStore = writable(null);
const store = (() => {
    return {
        subscribe: innerStore.subscribe,
        set: (d) => {
            if (d === null) {
                innerStore.set(null);
                value = d;
            }
            else if (d.getTime() !== $innerStore?.getTime()) {
                innerStore.set(d);
                value = d;
            }
        },
    };
})();
/** Date value */
export let value = null;
$: store.set(value);
/** The earliest value the user can select */
export let min = new Date(defaultDate.getFullYear() - 20, 0, 1);
/** The latest value the user can select */
export let max = new Date(defaultDate.getFullYear(), 11, 31, 23, 59, 59, 999);
/** Placeholder text to show when input field is empty */
export let placeholder = '2020-12-31 23:00:00';
/** Whether the text is valid */
export let valid = true;
/** Disable the input **/
export let disabled = false;
/** Format string */
export let format = 'yyyy-MM-dd HH:mm:ss';
let formatTokens = createFormat(format);
$: formatTokens = createFormat(format);
/** Locale object for internationalization */
export let locale = {};
function valueUpdate(value, formatTokens) {
    text = toText(value, formatTokens);
}
$: valueUpdate($store, formatTokens);
export let text = toText($store, formatTokens);
let textHistory = [text, text];
$: textHistory = [textHistory[1], text];
function textUpdate(text, formatTokens) {
    if (text.length) {
        const result = parse(text, formatTokens, $store);
        if (result.date !== null) {
            valid = true;
            store.set(result.date);
        }
        else {
            valid = false;
        }
    }
    else {
        valid = true; // <-- empty string is always valid
        // value resets to null if you clear the field
        if (value) {
            value = null;
            store.set(null);
        }
    }
}
$: textUpdate(text, formatTokens);
function input(e) {
    if (e instanceof InputEvent &&
        e.inputType === 'insertText' &&
        typeof e.data === 'string' &&
        text === textHistory[0] + e.data) {
        // check for missing punctuation, and add if there is any
        let result = parse(textHistory[0], formatTokens, $store);
        if (result.missingPunctuation !== '' && !result.missingPunctuation.startsWith(e.data)) {
            text = textHistory[0] + result.missingPunctuation + e.data;
        }
    }
}
/** Whether the date popup is visible */
export let visible = false;
/** Close the date popup when a date is selected */
export let closeOnSelection = false;
/** Wait with updating the date until a date is selected */
export let browseWithoutSelecting = false;
// handle on:focusout for parent element. If the parent element loses
// focus (e.g input element), visible is set to false
function onFocusOut(e) {
    if (e?.currentTarget instanceof HTMLElement &&
        e.relatedTarget &&
        e.relatedTarget instanceof Node &&
        e.currentTarget.contains(e.relatedTarget)) {
        return;
    }
    else {
        visible = false;
    }
}
function keydown(e) {
    if (e.key === 'Escape' && visible) {
        visible = false;
        e.preventDefault();
        // When the date picker is open, we prevent 'Escape' from propagating,
        // so for example a parent modal won't be closed
        e.stopPropagation();
    }
    else if (e.key === 'Enter') {
        visible = !visible;
        e.preventDefault();
    }
}

function onSelect(e) {
    dispatch('select', e.detail);
    if (closeOnSelection) {
      visible = false;
    }
}
</script>

<div class="date-time-field" on:focusout={onFocusOut} on:keydown={keydown}>
  <input class="input-calendar" readonly type="text"
    class:invalid={!valid}
    bind:value={text}
    {placeholder}
    {disabled}
    on:focus={() => (visible = true)}
    on:mousedown={() => (visible = true)}
    on:input={input}
  />
  <div class="calendar-icon">
    <img src="../images/calendar-icon.svg" alt="calendar icon">
  </div>

  {#if visible && !disabled}
    <div class="picker" class:visible transition:fly={{ duration: 80, easing: cubicInOut, y: -5 }}>
      <DateTimePicker
        on:focusout={onFocusOut}
        on:select={onSelect}
        bind:value={$store}
        {min}
        {max}
        {locale}
        {browseWithoutSelecting}
      />
    </div>
  {/if}
</div>

<style>

.date-time-field {
  position:relative;
}

.input-calendar {
  cursor: pointer;

  display: flex;
  justify-content: space-between;

  padding: 12px 11px 12px 16px;
  font-weight: 400;
  font-size: 1rem;
  line-height: 1rem;


  background-color: transparent;
  width: 100%;
  border: none;
  transition: all 80ms cubic-bezier(0.4, 0, 0.2, 1);

  outline: white;
}

.calendar-icon {
  position: absolute;
  bottom:10px;
  right:12px;
  z-index: -1;
}

input:disabled {
  opacity: 0.5;
}

.invalid {
  border: 1px solid rgba(249, 47, 114, 0.5);
  background-color: rgba(249, 47, 114, 0.1);
}
.invalid:focus {
  border-color: #f92f72;
  box-shadow: 0px 0px 0px 2px rgba(249, 47, 114, 0.5);
}

.picker {
  display: none;
  position: absolute;
  margin-top: 1px;
  z-index: 10;
}
.picker.visible {
  display: block;
}</style>
