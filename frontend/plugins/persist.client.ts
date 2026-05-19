export default defineNuxtPlugin(() => {
  const cartStore = useCartStore()
  const checkoutStore = useCheckoutStore()

  cartStore.hydrate()
  checkoutStore.hydrate()
})
