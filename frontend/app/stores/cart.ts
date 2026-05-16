import { defineStore } from 'pinia'

export interface CartItem {
  menu_item_id: number
  name: string
  quantity: number
  price: number
  restaurant_id: number
  restaurant_name?: string
}

export const useCartStore = defineStore('cart', {
  state: () => ({
    items: [] as CartItem[],
    restaurant_id: null as number | null,
    restaurant_name: '' as string,
  }),

  getters: {
    count: (state) => state.items.reduce((s, i) => s + i.quantity, 0),
    subtotal: (state) => state.items.reduce((s, i) => s + i.price * i.quantity, 0),
  },

  actions: {
    init() {
      if (import.meta.client) {
        const saved = localStorage.getItem('cart')
        if (saved) {
          const data = JSON.parse(saved)
          this.items = data.items ?? []
          this.restaurant_id = data.restaurant_id ?? null
          this.restaurant_name = data.restaurant_name ?? ''
        }
      }
    },

    _save() {
      if (import.meta.client) {
        localStorage.setItem('cart', JSON.stringify({
          items: this.items,
          restaurant_id: this.restaurant_id,
          restaurant_name: this.restaurant_name,
        }))
      }
    },

    addItem(item: CartItem): boolean {
      if (this.restaurant_id && this.restaurant_id !== item.restaurant_id) {
        return false
      }
      this.restaurant_id = item.restaurant_id
      this.restaurant_name = item.restaurant_name ?? ''
      const existing = this.items.find((i) => i.menu_item_id === item.menu_item_id)
      if (existing) {
        existing.quantity += item.quantity
      } else {
        this.items.push({ ...item })
      }
      this._save()
      return true
    },

    removeItem(menu_item_id: number) {
      this.items = this.items.filter((i) => i.menu_item_id !== menu_item_id)
      if (this.items.length === 0) {
        this.restaurant_id = null
        this.restaurant_name = ''
      }
      this._save()
    },

    updateQty(menu_item_id: number, qty: number) {
      const item = this.items.find((i) => i.menu_item_id === menu_item_id)
      if (item) {
        item.quantity = qty
        if (qty <= 0) this.removeItem(menu_item_id)
        else this._save()
      }
    },

    clear() {
      this.items = []
      this.restaurant_id = null
      this.restaurant_name = ''
      if (import.meta.client) localStorage.removeItem('cart')
    },
  },
})
