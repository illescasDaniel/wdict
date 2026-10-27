const runtimeURL = "/wails/runtime.js";
let runtime;
export async function call(method, ...args) {
  if (!runtime) runtime = await import(/* @vite-ignore */ runtimeURL);
  return runtime.Call.ByName("main.App." + method, ...args);
}
