import { PayPalScriptProvider, PayPalButtons } from "@paypal/react-paypal-js"
import { PaymentEventSessionCreatedData } from "../contracts"
import { useRouter } from "next/navigation"

interface PayPalPaymentButtonProps {
  paymentSession: PaymentEventSessionCreatedData
}

export const PayPalPaymentButton = ({ paymentSession }: PayPalPaymentButtonProps) => {
  const router = useRouter()
  const clientId = process.env.NEXT_PUBLIC_PAYPAL_CLIENT_ID

  if (!clientId) {
    return (
      <div className="w-full p-3 bg-red-100 text-red-700 text-xs rounded border border-red-300">
        NEXT_PUBLIC_PAYPAL_CLIENT_ID is missing in your environment configuration
      </div>
    )
  }

  return (
    <PayPalScriptProvider
      options={{
        clientId: clientId,
        currency: paymentSession.currency ? paymentSession.currency.toUpperCase() : "USD",
        intent: "capture",
      }}
    >
      <PayPalButtons
        style={{ layout: "vertical" }}
        createOrder={() => {
          return Promise.resolve(paymentSession.sessionID)
        }}
        onApprove={async (data, actions) => {
          console.log("PayPal payment approved:", data.orderID)

          // 1. Capture order payment
          if (actions.order) {
            await actions.order.capture()
          }

          // 2. Redirect to triggers the green checkmark screen in app/page.tsx
          router.push("/?payment=success")
        }}
        onError={(err) => {
          console.error("PayPal SDK Error:", err)
        }}
      />
    </PayPalScriptProvider>
  )
}