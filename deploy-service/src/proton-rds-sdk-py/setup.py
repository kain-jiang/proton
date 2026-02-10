"""Setup Script for DBUtilsX"""

import warnings
import os
import shutil

try:
    from setuptools import setup
except ImportError:
    from distutils.core import setup
from sys import version_info

import platform
machine = platform.machine()
major, minor = version_info[:2]
py_version = f'{major}{minor}'
dm_so_file = f'rdsdriver/dm/dmPython.cpython-{py_version}-{machine}-linux-gnu.so'
if os.path.exists(dm_so_file):
    shutil.move(dm_so_file, 'rdsdriver/dm/dmPython.so')
else:
    shutil.move(f'rdsdriver/dm/dmPython.cpython-38-{machine}-linux-gnu.so', 'rdsdriver/dm/dmPython.so')

py_version = version_info[:2]
if py_version != (2, 7) and not (3, 5) <= py_version < (4, 0):
    raise ImportError('Python %d.%d is not supported by DBUtils.' % py_version)

warnings.filterwarnings('ignore', 'Unknown distribution option')

__version__ = '1.4.2'

readme = open('README.md').read()

setup(
    name='DBUtilsX',
    version=__version__,
    description='Database connections for multi-threaded environments.',
    project_urls={
        'Source Code':
            'https://devops.aishu.cn/AISHUDevOps/ICT/_git/proton-rds-sdk-py'},
    platforms=['any'],
    license='MIT License',
    packages=['dbutilsx', 'dbutilsx.dbutils']
)

setup(
    name='RDSDriver',
    version=__version__,
    description='RDS DB API 2.0 driver.',
    project_urls={
        'Source Code':
            'https://devops.aishu.cn/AISHUDevOps/ICT/_git/proton-rds-sdk-py'},
    platforms=['x86_64', 'aarch64'],
    license='MIT License',
    packages=['rdsdriver','rdsdriver.dm','rdsdriver.mysql', 'rdsdriver.kingbase'],
    include_package_data=True,
    python_requires=">=3.0, <4",
    zip_safe=False,
    install_requires = [
        "pymysql",
        "psycopg2"
    ]
)

if os.path.exists('./build'):
    shutil.rmtree('./build')
if os.path.exists('./dist'):
    shutil.rmtree('./dist')
if os.path.exists('./DBUtilsX.egg-info'):
    shutil.rmtree('./DBUtilsX.egg-info')
if os.path.exists('./RDSDriver.egg-info'):
    shutil.rmtree('./RDSDriver.egg-info')



