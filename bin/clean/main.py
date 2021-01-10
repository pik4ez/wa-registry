#!/usr/bin/env python

import os
import yaml

from registry import Registry
from typing import Dict, List


DEFAULT_LIMIT = 10

REGISTRY_URL = os.environ.get('REGISTRY_URL')
USERNAME = os.environ.get('REGISTRY_USER')
PASSWORD = os.environ.get('REGISTRY_PASSWORD')
CONFIG_PATH = os.environ.get('CONFIG')


def get_config(path):
    with open(path, 'r') as stream:
        return yaml.safe_load(stream)


def get_limit_by_repo(cfg: dict, repo_name: str) -> Dict[str, int]:
    return cfg.get(repo_name, DEFAULT_LIMIT)


def get_tags_to_delete(tags: List[str], limit: int) -> List[str]:
    tags = [t for t in tags if t != 'latest']
    return tags[limit:]


client = Registry(
    REGISTRY_URL,
    username=USERNAME,
    password=PASSWORD)

cfg = get_config(CONFIG_PATH)

repos = client.list_repos()

for repo in repos:
    limit = get_limit_by_repo(cfg, repo)
    tags = client.list_tags(repo)
    print(f'tags {tags}')
    tags_to_delete = get_tags_to_delete(tags, limit)
    for tag in tags_to_delete:
        print(f'deleting tag {repo}/{tag}')
        digest = client.get_digest_by_tag(repo, tag)
        code = client.delete_tag(repo, digest)
        if code != 202:
            raise Exception(
                f'Failed to delete tag {tag} in repo {repo}.' +
                f' Digest {digest}.' +
                f' Return code {code}.'
            )
