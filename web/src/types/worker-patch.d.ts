// Type patch for Vite's importGlob.d.ts Worker dependency
// This provides the Worker type when WebWorker lib is not included

declare interface Worker extends EventTarget {
  onerror: ((this: Worker, ev: ErrorEvent) => any) | null;
  onmessage: ((this: Worker, ev: MessageEvent<any>) => any) | null;
  onmessageerror: ((this: Worker, ev: MessageEvent<any>) => any) | null;
  postMessage(message: any, transfer?: Transferable[]): void;
  postMessage(message: any, options?: StructuredSerializeOptions): void;
  terminate(): void;
  addEventListener<K extends keyof WorkerEventMap>(
    type: K,
    listener: (this: Worker, ev: WorkerEventMap[K]) => any,
    options?: boolean | AddEventListenerOptions
  ): void;
  addEventListener(
    type: string,
    listener: EventListenerOrEventListenerObject,
    options?: boolean | AddEventListenerOptions
  ): void;
  removeEventListener<K extends keyof WorkerEventMap>(
    type: K,
    listener: (this: Worker, ev: WorkerEventMap[K]) => any,
    options?: boolean | EventListenerOptions
  ): void;
  removeEventListener(
    type: string,
    listener: EventListenerOrEventListenerObject,
    options?: boolean | EventListenerOptions
  ): void;
}

declare interface WorkerEventMap {
  error: ErrorEvent;
  message: MessageEvent<any>;
  messageerror: MessageEvent<any>;
}

declare var Worker: {
  prototype: Worker;
  new (url: string | URL, options?: WorkerOptions): Worker;
};

declare interface WorkerOptions {
  credentials?: RequestCredentials;
  name?: string;
  type?: WorkerType;
}

declare type WorkerType = "classic" | "module";

declare interface StructuredSerializeOptions {
  transfer?: Transferable[];
}
