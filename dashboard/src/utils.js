export function buildPayload(command, data) {
  return JSON.stringify({"command":command, "data":data})
}

