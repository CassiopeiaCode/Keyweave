export const manifest = {
  id: "oauth-refresh-example",
  apiVersion: "ptpg.template/v1",
  protocols: ["openai.responses"],
  defaultDecrease: 1,
  capabilities: {
    credentialBoundNetwork: true,
    mutateInstance: true,
    setSecrets: true
  }
};

async function ensureAccessToken(ctx) {
  const expiresAt = Number(ctx.fields.accessTokenExpiresAt ?? 0);
  const current = await ctx.secrets.get("accessToken");

  if (current && expiresAt > ctx.clock.now() + 60_000) {
    return current;
  }

  const refresh = await ctx.secrets.get("refreshToken");
  if (!refresh) throw new Error("refreshToken missing");

  const resp = await ctx.http.request({
    url: String(ctx.fields.tokenEndpoint),
    method: "POST",
    headers: {
      "content-type": "application/x-www-form-urlencoded"
    },
    body:
      "grant_type=refresh_token&refresh_token=" +
      encodeURIComponent(refresh.reveal())
  });

  if (!resp.json) throw new Error("HostResponse.json not available");
  const payload = await resp.json();

  await ctx.instance.setSecret("accessToken", payload.access_token);
  await ctx.instance.patch({
    customFields: {
      accessTokenExpiresAt:
        ctx.clock.now() +
        Number(payload.expires_in ?? 3600) * 1000
    }
  });

  return ctx.secrets.get("accessToken");
}

export async function execute(ctx) {
  const token = await ensureAccessToken(ctx);
  const body = await ctx.request.json();
  body.model = ctx.route.upstreamModel;

  return ctx.http.request({
    url: String(ctx.fields.baseUrl).replace(/\/$/, "") + ctx.request.path,
    method: ctx.request.method,
    headers: {
      "content-type": "application/json",
      authorization: `Bearer ${token.reveal()}`
    },
    body: JSON.stringify(body)
  });
}
