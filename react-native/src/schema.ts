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

// Protocol v2 of ../../web-wasm/wasm-template.html: the page never forwards
// console output or error details, only these four message types.
export const wasmMessageSchema = z.preprocess(
  parseJsonPreprocessor,
  z.discriminatedUnion('type', [
    z.object({ type: z.literal('ready'), version: z.literal(2) }),
    z.object({ type: z.literal('executeResult'), executeId: z.string(), result: z.unknown() }),
    z.object({ type: z.literal('executeError'), executeId: z.string() }),
    z.object({ type: z.literal('error') }),
  ]),
);

export type WasmMessage = z.infer<typeof wasmMessageSchema>;
