import { AlertCircle, CheckCircle2, LoaderCircle, WifiOff } from "lucide-react";

import { Button } from "@/components/ui/button";

type OperationFeedbackProps = {
  status: "loading" | "success" | "error" | "disconnected";
  message: string;
  onDismiss?: () => void;
  onRetry?: () => void;
};

export function OperationFeedback({
  status,
  message,
  onDismiss,
  onRetry,
}: OperationFeedbackProps) {
  const Icon =
    status === "loading"
      ? LoaderCircle
      : status === "success"
        ? CheckCircle2
        : status === "disconnected"
          ? WifiOff
          : AlertCircle;

  return (
    <div
      className={`operation-feedback ${status}`}
      role={status === "error" ? "alert" : "status"}
      aria-live={status === "error" ? "assertive" : "polite"}
    >
      <Icon aria-hidden="true" />
      <span>{message}</span>
      <div className="operation-feedback-actions">
        {onRetry ? (
          <Button variant="outline" size="sm" onClick={onRetry}>
            Retry
          </Button>
        ) : null}
        {onDismiss ? (
          <Button variant="ghost" size="sm" onClick={onDismiss}>
            Dismiss
          </Button>
        ) : null}
      </div>
    </div>
  );
}
