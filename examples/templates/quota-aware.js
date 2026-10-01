export const manifest = {
  id: "quota-aware-example",
  apiVersion: "ptpg.template/v1",
  protocols: ["openai.chat"],
  defaultDecrease: 5,
  capabilities: {
    credentialBoundNetwork: true,
    mutateInstance: true
  }
};

export function getAvailability(ctx) {
  const remaining = Number(
    ctx.fields.quotaRemaining ?? Number.MAX_SAFE_INTEGER
  );
  const resetAt = Number(ctx.fields.quotaResetAt ?? 0);

  if (remaining <= 0 && resetAt > ctx.clock.now()) {
    return -1_000_000_000;
  }

  const base = Number(ctx.instance.availability ?? 100);
  if (remaining < 1000) return base - 50;
  return base;
}

export async function afterResponse(ctx, response) {
  const remaining = response.headers.get("x-ratelimit-remaining");
  const reset = response.headers.get("x-ratelimit-reset-ms");

  if (remaining !== undefined || reset !== undefined) {
    await ctx.instance.patch({
      customFields: {
        ...(remaining !== undefined
          ? { quotaRemaining: Number(remaining) }
          : {}),
        ...(reset !== undefined
          ? { quotaResetAt: Number(reset) }
          : {})
      }
    });
  }
}
