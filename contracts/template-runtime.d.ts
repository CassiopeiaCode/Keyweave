import type { ModelDefinition, ProtocolId } from "./domain";

export interface TemplateManifest {
  id: string;
  apiVersion: "ptpg.template/v1";
  protocols: ProtocolId[];
  defaultDecrease?: number;
  capabilities?: {
    credentialBoundNetwork?: boolean;
    directNetwork?: boolean;
    mutateInstance?: boolean;
    mutateTemplate?: boolean;
    setSecrets?: boolean;
  };
}

export interface SessionMetadata {
  id?: string;
  conversationId?: string;
  promptCacheKey?: string;
  parentSessionId?: string;
  clientRequestId?: string;
  derivedHash?: string;
}

export interface HeaderMap {
  get(name: string): string | undefined;
  set(name: string, value: string): void;
  delete(name: string): void;
  entries(): Array<[string, string]>;
}

export interface TemplateRequest {
  requestId: string;
  protocol: ProtocolId;
  model: string;
  upstreamModel: string;
  stream: boolean;
  method: string;
  path: string;
  query: string;
  headers: HeaderMap;
  session?: SessionMetadata;

  json<T = unknown>(): Promise<T>;
  text(): Promise<string>;
}

export interface SecretString {
  readonly __secretBrand: unique symbol;
  reveal(): string;
}

export interface SecretAccessor {
  get(name: string): Promise<SecretString | undefined>;
}

export interface HostReadableStream {
  [Symbol.asyncIterator](): AsyncIterator<Uint8Array>;
}

export interface HostResponse {
  status: number;
  headers: HeaderMap;
  body: HostReadableStream;
  json?<T = unknown>(): Promise<T>;
  text?(): Promise<string>;
}

export interface HostHttp {
  request(input: {
    url: string;
    method?: string;
    headers?: Record<string, string>;
    body?: string | Uint8Array | HostReadableStream;
    timeoutMs?: number;
    proxyMode?: "instance" | "direct";
  }): Promise<HostResponse>;

  websocket?(
    url: string,
    options?: Record<string, unknown>
  ): Promise<unknown>;
}

export interface AtomicStateApi {
  increment(path: string, delta: number): Promise<number>;
  max(path: string, value: number): Promise<number>;
}

export interface MutableInstanceApi {
  readonly id: string;
  readonly templateId: string;
  readonly availability: number;
  readonly activeRequests: number;

  patch(input: {
    availability?: number;
    customFields?: Record<string, unknown>;
    state?: Record<string, unknown>;
  }): Promise<void>;

  setSecret?(name: string, value: string): Promise<void>;
  atomic: AtomicStateApi;
}

export interface MutableTemplateApi {
  readonly id: string;
  readonly version: number;
  patchState(statePatch: Record<string, unknown>): Promise<void>;
  atomic: AtomicStateApi;
}

export interface TemplateContext {
  request: TemplateRequest;
  fields: Readonly<Record<string, unknown>>;
  secrets: SecretAccessor;

  instance: MutableInstanceApi;
  template: MutableTemplateApi;

  route: {
    groupId: string;
    clientKeyId: string;
    requestedModel: string;
    upstreamModel: string;
  };

  runtime: {
    attempt: number;
    activeRequests: number;
    downstreamStarted: boolean;
  };

  http: HostHttp;

  log: {
    debug(...args: unknown[]): void;
    info(...args: unknown[]): void;
    warn(...args: unknown[]): void;
    error(...args: unknown[]): void;
  };

  clock: {
    now(): number;
  };
}

export interface ScheduleHints {
  hardMaxConcurrency?: number;
  attributes?: Record<string, unknown>;
}

export type ErrorAction =
  | { action: "stop" }
  | {
      action: "retry";
      target: "same" | "another";
      cooldownMs?: number;
    };

export declare const manifest: TemplateManifest;

export declare function onSystemStart?(
  ctx: TemplateContext
): Promise<void>;

export declare function onInstanceStart?(
  ctx: TemplateContext
): Promise<void>;

export declare function supports?(
  ctx: TemplateContext
): boolean | Promise<boolean>;

export declare function getModels?(
  ctx: TemplateContext
): ModelDefinition[] | Promise<ModelDefinition[]>;

export declare function getAvailability?(
  ctx: TemplateContext
): number | Promise<number>;

export declare function getDecrease?(
  ctx: TemplateContext
): number | Promise<number>;

export declare function getScheduleHints?(
  ctx: TemplateContext
): ScheduleHints | Promise<ScheduleHints>;

export declare function getProxy?(
  ctx: TemplateContext
): string | null | Promise<string | null>;

export declare function beforeRequest?(
  ctx: TemplateContext
): Promise<void>;

export declare function execute?(
  ctx: TemplateContext
): Promise<HostResponse>;

export declare function afterResponse?(
  ctx: TemplateContext,
  response: HostResponse
): Promise<void>;

export declare function onError?(
  ctx: TemplateContext,
  error: unknown
): Promise<ErrorAction>;

export declare function onFinish?(
  ctx: TemplateContext
): Promise<void>;
