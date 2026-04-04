import request from '@/utils/request'

const basePath = '/api/v1/codegen_consoles'

export function listcodegen_consoles(params = {}) {
  return request({
    url: basePath,
    method: 'get',
    params,
  })
}

export function getCodegenConsole(id) {
  return request({
    url: basePath + '/' + id,
    method: 'get',
  })
}

export function createCodegenConsole(data) {
  return request({
    url: basePath,
    method: 'post',
    data,
  })
}

export function updateCodegenConsole(id, data) {
  return request({
    url: basePath + '/' + id,
    method: 'put',
    data,
  })
}

export function deleteCodegenConsole(id) {
  return request({
    url: basePath + '/' + id,
    method: 'delete',
  })
}
