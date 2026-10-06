"""Проверки переносимости сохраняемых JSON reports."""
import unittest
from pathlib import Path
from urllib.parse import quote

from report_paths import portable_report


class ReportPathsTest(unittest.TestCase):
    """Пути окружения удаляются без изменения диагностик и validity."""

    def test_nested_diagnostics(self):
        root = Path('/') / 'Users' / 'example' / 'project'
        report = {'valid': False, 'matches': 2, 'diagnostics': [
            {'file': str(root / 'testdata/model.cml'), 'line': 4, 'code': 'syntax'}]}
        self.assertEqual(portable_report(report, root), {
            'valid': False, 'matches': 2, 'diagnostics': [
                {'file': 'testdata/model.cml', 'line': 4, 'code': 'syntax'}]})

    def test_reference_and_encoded_paths(self):
        root = Path('/') / 'Users' / 'example' / 'project'
        library = root.parent / 'Tool Directory' / 'library'
        output = f'file:{library}/lib.jar; file:{quote(str(library))}/lib.jar'
        self.assertEqual(portable_report(output, root, [library]),
                         'file:reference/lib.jar; file:reference/lib.jar')

    def test_other_user_paths(self):
        root = Path('/workspace')
        path = Path('/') / 'home' / 'example' / 'tools' / 'lib.jar'
        self.assertEqual(portable_report(str(path), root), 'reference/lib.jar')
        self.assertEqual(portable_report(['testdata/model.cml', True, None], root),
                         ['testdata/model.cml', True, None])


if __name__ == '__main__':
    unittest.main()
