import { z } from 'zod';

const parseJsonPreprocessor = (value: unknown, ctx: z.RefinementCtx) => {
  if (typeof value !== 'string') {
    return value;
  }

  try {
    return JSON.parse(value);
  } catch {
    ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Invalid JSON payload from WASM WebView' });
    return z.NEVER;
  }
};

export const wasmMessageSchema = z.preprocess(
  parseJsonPreprocessor,
  z.discriminatedUnion('type', [
    z.object({ type: z.literal('ready') }),
    z.object({ type: z.literal('log'), message: z.union([z.string(), z.string().array()]) }),
    z.object({ type: z.literal('warn'), message: z.union([z.string(), z.string().array()]) }),
    z.object({ type: z.literal('error'), message: z.union([z.string(), z.string().array()]) }),
    z.object({ type: z.literal('executeResult'), executeId: z.string(), result: z.unknown() }),
    z.object({ type: z.literal('executeError'), executeId: z.string(), message: z.string() }),
  ]),
);

export type WasmMessage = z.infer<typeof wasmMessageSchema>;
