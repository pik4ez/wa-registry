import requests

from typing import List


class Registry:
    def __init__(self, url: str, username: str, password: str):
        self._url = '{}/v2'.format(url)
        self._username = username
        self._password = password

        self._headers = {
            'Accept': 'application/vnd.docker.distribution.manifest.v2+json',
        }

    def list_repos(self) -> List[str]:
        url = '{}/_catalog'.format(self._url)
        r = requests.get(
            url,
            auth=(
                self._username,
                self._password),
            headers=self._headers)
        resp = r.json()
        return resp['repositories']

    def list_tags(self, repo: str) -> List[str]:
        url = '{}/{}/tags/list'.format(self._url, repo)
        r = requests.get(
            url,
            auth=(
                self._username,
                self._password),
            headers=self._headers)
        resp = r.json()
        return resp['tags']

    def get_digest_by_tag(self, repo: str, tag: str) -> str:
        url = '{}/{}/manifests/{}'.format(self._url, repo, tag)
        r = requests.get(
            url,
            auth=(
                self._username,
                self._password),
            headers=self._headers)
        return r.headers.get('docker-content-digest')

    def delete_tag(self, repo: str, digest: str) -> int:
        url = '{}/{}/manifests/{}'.format(self._url, repo, digest)
        r = requests.delete(
            url,
            auth=(
                self._username,
                self._password),
            headers=self._headers)
        return r.status_code
