export const manifest = {
  id: "generic-openai",
  apiVersion: "ptpg.template/v1",
  protocols: ["openai.chat", "openai.responses"],
  defaultDecrease: 1,
  capabilities: {
    credentialBoundNetwork: true,
    mutateInstance: true
  }
};

export function getAvailability(ctx) {
  return Number(ctx.instance.availability ?? 100);
}

export function getDecrease(ctx) {
  return Number(ctx.fields.decrease ?? 1);
}

export function getScheduleHints(ctx) {
  const max = Number(ctx.fields.maxConcurrency ?? 0);
  return max > 0 ? { hardMaxConcurrency: max } : {};
}

export async function execute(ctx) {
  const key = await ctx.secrets.get("key");
  if (!key) throw new Error("credential key is missing");

  const baseUrl = String(ctx.fields.baseUrl ?? "");
  if (!baseUrl) throw new Error("baseUrl is required");

  const body = await ctx.request.json();
  body.model = ctx.route.upstreamModel;

  const headers = Object.fromEntries(ctx.request.headers.entries());
  delete headers.authorization;
  delete headers.Authorization;
  delete headers["x-api-key"];

  headers.authorization = `Bearer ${key.reveal()}`;
  headers["content-type"] = "application/json";

  return ctx.http.request({
    url: baseUrl.replace(/\/$/, "") + ctx.request.path,
    method: ctx.request.method,
    headers,
    body: JSON.stringify(body)
  });
}

export async function onError(ctx, error) {
  const status = Number(error?.upstreamStatus ?? error?.status ?? 0);

  if ([408, 429, 500, 502, 503, 504].includes(status)) {
    return { action: "retry", target: "another" };
  }

  if ([401, 403].includes(status)) {
    await ctx.instance.patch({
      availability: -1000000
    });
  }

  return { action: "stop" };
}
