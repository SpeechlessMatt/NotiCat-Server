# Author: Czy_4201b <speechlessmatt@qq.com>
# Created: 2026-02-03

import os
import importlib
from .base import BaseClient, CLIENT_REGISTER

pkg_dir = os.path.dirname(__file__)

for module in os.listdir(pkg_dir):
    if module.endswith(".py") and module not in ["__init__.py", "base.py"]:
        # remove '.py'
        name = module[:-3] 
        importlib.import_module(f".{name}", package=__package__)

clients = CLIENT_REGISTER

__all__ = ["clients", "BaseClient"]
