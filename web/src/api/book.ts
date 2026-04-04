import request from '@/utils/request'

const basePath = '/api/v1/books'

export function listbooks(params = {}) {
  return request({
    url: basePath,
    method: 'get',
    params,
  })
}

export function getBook(id) {
  return request({
    url: basePath + '/' + id,
    method: 'get',
  })
}

export function createBook(data) {
  return request({
    url: basePath,
    method: 'post',
    data,
  })
}

export function updateBook(id, data) {
  return request({
    url: basePath + '/' + id,
    method: 'put',
    data,
  })
}

export function deleteBook(id) {
  return request({
    url: basePath + '/' + id,
    method: 'delete',
  })
}
