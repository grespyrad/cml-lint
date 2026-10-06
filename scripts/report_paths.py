"""Удаление локальных путей из сохраняемых отчётов совместимости."""
import re
from pathlib import Path
from urllib.parse import quote


def portable_report(value, root, extra_roots=()):
    """Сохраняет структуру отчёта и заменяет пути локального окружения."""
    if isinstance(value, dict):
        return {key: portable_report(item, root, extra_roots) for key, item in value.items()}
    if isinstance(value, list):
        return [portable_report(item, root, extra_roots) for item in value]
    if isinstance(value, str):
        result = value.replace(str(root) + '/', '').replace(quote(str(root)) + '/', '')
        for location in extra_roots:
            result = result.replace(str(location), 'reference').replace(quote(str(location)), 'reference')
        return re.sub(r'/(?:Users|home)/[^\s\"\'()<>]+',
                      lambda match: 'reference/' + Path(match.group()).name, result)
    return value
