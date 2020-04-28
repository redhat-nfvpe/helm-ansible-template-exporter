#!/usr/bin/python

"""
Module responsible for bridging Ansible Filters with Golang template functions.  This is a work in progress, and
provided with the understanding that type-mapping is not really supported.  That is, this tool expects string I/O, and
cannot handle Golang template or Sprig functions which return more complex data types (such as maps).

Here is an example how to invoke the go "shuffle" template function from an Ansible Playbook:
- debug:
    msg:
    - "{{ 'aaaaaaabbbbbbbbbccccccccddddddddd' | invoke_go_tf('shuffle') }}"

Utilities to simplify this syntax are provided by the "_generic_wrapped_tf" method.  Two example implementations are
"trim" and "contains".  Note:  these functions are commented out of the "filters" method, so they are not enabled by
default.
"""

import logging
import os
import shutil
import subprocess
from typing import AnyStr
from typing import Dict, Any, List


class FilterModule(object):
    """
    FilterModule is an abstraction responsible for providing Golang Template and Sprig Function functionality through
    an Ansible Playbook Filter.  This class will contain functionality that attempts to directly mimic existing Golang
    template functionality to provide a 1:1 conversion when possible.
    """

    # Filter name constants
    CONTAINS_KEY: str = "contains"
    INVOKE_GO_TF__KEY: str = "invoke_go_tf"
    TRIM__KEY: str = "trim"

    # Other constants
    ANSIBLE_ARGUMENT_ONE: int = 0
    ANSIBLE_ARGUMENT_TWO: int = 1
    ANSIBLE_ARGUMENT_THREE: int = 2
    DEFAULT_ENCODING: str = "UTF-8"
    GO_BINARY_NAME: str = "go"
    GO_RUN_KEYWORD: str = "run"
    SPRIG_CONDUIT_FILENAME: str = "main.go"
    SPRIG_CONDUIT_PATH: AnyStr = os.path.join(os.path.dirname(__file__), SPRIG_CONDUIT_FILENAME)
    ZERO_ARGS: int = 0

    def __init__(self):
        super(FilterModule, self).__init__()
        # set up some defaults
        self._log = self._setup_logging()
        self._go_binary = self._establish_go_binary()

    def _setup_logging(self) -> logging.Logger:
        """
        Sets up a logging.Logger equipped with a logging.StreamHandler, which will output log messages to the console.
        :return: A preconfigured logging.Logger
        """
        self._log = logging.getLogger(__name__)
        self._log.setLevel(logging.INFO)
        self._log.addHandler(logging.StreamHandler())
        return self._log

    def _establish_go_binary(self):
        """
        Establish the go binary location within the system $PATH
        :return: the file path location of the go binary
        """
        go_binary = shutil.which(FilterModule.GO_BINARY_NAME)
        if go_binary is None:
            self._log.error("Couldn't locate go binary in $PATH")
        self._log.debug("Found go binary: %s", go_binary)
        return go_binary

    def filters(self) -> Dict[str, Any]:
        """
        Returns a list of exposed filters to the Ansible runtime.  Since Ansible provides no base class abstraction,
        this method is assumed to be present, though it is not contractually obligated.
        :return: a dict with keys that are the filter keywords and values are the filter implementation
        """
        # Uncomment those which make sense.
        return {
            # FilterModule.CONTAINS_KEY: self.contains,
            FilterModule.INVOKE_GO_TF__KEY: self.invoke_go_tf,
            # FilterModule.TRIM__KEY: self.trim,
        }

    def trim(self, *args) -> str:
        """
        A conduit between Ansible Filter and Go Sprig's "trim" function.  This invokes a Go subprocess with the
        corresponding arguments and return the raw string output.

        To enable, see "filters" function.

        :param args: Ansible Playbook arguments
        :return: subprocess output after rendering the Go template
        """
        return self._generic_wrapped_tf(FilterModule.TRIM__KEY, args)

    def contains(self, *args) -> str:
        """
        A conduit between Ansible Filter and Go Sprig's "contains" function.  This invokes a Go subprocess with the
        corresponding arguments and return the raw string output.

        To enable, see "filters" function.

        :param args: Ansible Playbook arguments
        :return: subprocess output after rendering the Go template
        """
        return self._generic_wrapped_tf(FilterModule.CONTAINS_KEY, args)

    @staticmethod
    def _first_argument_exists(args: [str]) -> bool:
        """
        Verifies that the first Ansible Argument exists.  Since Ansible Filters are invoked by piping, the first
        argument should always exist from this context;  this is considered a sanity check.
        :param args: Ansible Playbook Filter arguments
        :return: whether the first argument exists
        """
        return len(args) > FilterModule.ZERO_ARGS

    def _generic_wrapped_tf(self, func_name: str, *args: [str]) -> str:
        """
        Provides an easy way to invoke Go templating for any generic Go/Sprig Template Function.  See "trim" and
        "contains" for examples.
        :param func_name: the name of the go template function
        :param args: the Ansible Playbook Filter arguments
        :return: the output after rendering the Go template
        """
        # The first argument is guaranteed to exist.  Check anyway;  weirder things have happened.
        if not FilterModule._first_argument_exists(args=args):
            self._log.error("Ansible Playbook should guarantee at least one argument;  check the input playbook")

        # re-arrange arguments to expected format
        arguments = args[FilterModule.ANSIBLE_ARGUMENT_ONE]
        list_arguments = list(arguments)
        augmented_args = list()
        augmented_args.append(list_arguments[FilterModule.ANSIBLE_ARGUMENT_ONE])
        augmented_args.append(func_name)
        for i in range(FilterModule.ANSIBLE_ARGUMENT_TWO, len(list_arguments)):
            augmented_args.append(list_arguments[i])
        return self.invoke_go_tf(*augmented_args)

    def invoke_go_tf(self, *args):
        """
        invokes a generic go template function.
        :param args: Ansible Playbook Filter arguments
        :return: the output after rendering the Go template
        """
        return self._invoke(args)

    @staticmethod
    def _form_system_call(go_binary: str, func_name: str, first_arg: str, other_args: List[str]) -> List[str]:
        """
        Form the system call used to invoke the Go template shim.
        :param go_binary: the filename of the go binary
        :param func_name: the name of the sprig/Go template function
        :param first_arg: the first argument (the one before the pipe in Ansible filter)
        :param other_args: all other arguments to the Sprig function
        :return: the raw command array which can be passed to subprocess.Popen.
        """
        # re-arrange applicable arguments
        sys_call_list = list()
        sys_call_list.append(go_binary)
        sys_call_list.append(FilterModule.GO_RUN_KEYWORD)
        sys_call_list.append(FilterModule.SPRIG_CONDUIT_PATH)
        sys_call_list.append(func_name)
        # only shim in the first argument if it is not empty
        if first_arg != "":
            sys_call_list.append(first_arg)
        for arg in other_args:
            sys_call_list.append(str(arg))
        return sys_call_list

    def _invoke(self, args):
        """
        Invokes the Go subprocess with applicable arguments in order to resolve sprig function I/O.
        :param args: Ansible Playbook Filter arguments.
        :return: the output after rendering the Go template
        """
        self._log.info("_invoke%s", args)
        first_arg = args[FilterModule.ANSIBLE_ARGUMENT_ONE]
        func_name = args[FilterModule.ANSIBLE_ARGUMENT_TWO]
        other_args = args[FilterModule.ANSIBLE_ARGUMENT_THREE:]

        sys_call_list = FilterModule._form_system_call(self._go_binary, func_name, first_arg, other_args)
        self._log.info("Go System Call: %s", " ".join(sys_call_list))

        process = subprocess.Popen(sys_call_list, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        stdout_bytes, stderr_bytes = process.communicate()
        if stderr_bytes is not None and stderr_bytes.decode(FilterModule.DEFAULT_ENCODING) != "":
            self._log.error("Go invocation attempt failed! stderr: %s",
                            stderr_bytes.decode(FilterModule.DEFAULT_ENCODING))
        stdout = stdout_bytes.decode(FilterModule.DEFAULT_ENCODING)
        return stdout
