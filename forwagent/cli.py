import argparse
import logging
import sys
import os
from .common import CONF_DIR


def init_config(args):
    try:
        from .config import init

        init()
    except ImportError:
        print(
            "Missing dependencies for 'init', "
            "re-install with 'pip install forwagent[init]'"
        )
        sys.exit(1)


def exec_agent(args):
    if not os.path.isdir(CONF_DIR):
        print("Configuration directory missing, run 'forwagent init' to initialize.")
        sys.exit(1)

    from .agent import main

    main((args.interface, args.port))


def exec_server(args):
    if not os.path.isdir(CONF_DIR):
        print("Configuration directory missing, run 'forwagent init' to initialize.")
        sys.exit(1)

    try:
        from .server import main

        main((args.interface, args.port))
    except ImportError:
        print(
            "Missing dependencies for 'server', "
            "re-install with 'pip install forwagent[server]'"
        )
        sys.exit(1)


def get_parser():
    parser = argparse.ArgumentParser()
    parser.set_defaults(func=None)
    parser.add_argument("--verbose", action="store_true", help="print debug output")
    parser.add_argument("--log-file", help="write log output to file")

    subparsers = parser.add_subparsers(title="subcommands")

    init = subparsers.add_parser("init", help="initialize configuration")
    init.set_defaults(func=init_config)

    agent = subparsers.add_parser("agent", help="run the agent")
    agent.set_defaults(func=exec_agent)
    agent.add_argument(
        "--interface",
        default="127.0.0.1",
        help="network interface to connect to (default: 127.0.0.1)",
    )
    agent.add_argument(
        "--port",
        default=4711,
        type=int,
        help="network port to connect to (default: 4711)",
    )

    server = subparsers.add_parser("server", help="run the server")
    server.set_defaults(func=exec_server)
    server.add_argument(
        "--interface",
        default="127.0.0.1",
        help="network interface to listen on (default: 127.0.0.1)",
    )
    server.add_argument(
        "--port",
        default=4711,
        type=int,
        help="network port to listen on (default: 4711)",
    )

    return parser


def main():
    parser = get_parser()
    args = parser.parse_args()
    if args.func:
        if args.verbose:
            logging.basicConfig(
                datefmt="%Y-%m-%dT%H:%M:%S%z",
                filename=args.log_file,
                format="%(asctime)s [%(levelname)s] %(message)s",
                level=logging.DEBUG,
            )

        args.func(args)
    else:
        parser.print_help()
