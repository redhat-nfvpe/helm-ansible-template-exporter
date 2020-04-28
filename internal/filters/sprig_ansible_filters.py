#!/usr/bin/python

"""
Module stubbing out all Sprig Functions.  All function implementations just raise NotImplementedError.

The naming convention for these functions is "sprig_<sprigFunctionName>".  Nothing is done to convert Go Template
function names into Python function convention (i.e., conversion to snake case).  One should not name any additional
functions with the prefix "sprig_" unless they are adding an additional Sprig function implementation.  The "filters"
function returns keys without the "sprig_" prefix, so that Go Template functions that are referenced in Ansible
Playbooks do not require the "sprig_" prefix.  For example:

- debug:
    msg:
    - "{{ 'aaaaaaabbbbbbbbbccccccccddddddddd' | shuffle }}"
"""

import inspect


class FilterModule(object):
    """
    Defines stub methods to implement Go template functions.
    """

    REPLACE_ONCE = 1
    SPRIG_PREFIX = "sprig_"

    def __init__(self):
        super(FilterModule, self) .__init__()
        # Any function in this class with a name that starts with the prefix "sprig_" is considered an Ansible Filter
        # implementation.  As such, it is added to the private function map called self._ansible_filters
        full_func_map = inspect.getmembers(FilterModule, predicate=inspect.isfunction)
        self._ansible_filters = {}
        for func_name, func in full_func_map:
            if func_name.startswith(FilterModule.SPRIG_PREFIX):
                ansible_filter_key = func_name.replace(FilterModule.SPRIG_PREFIX, "", FilterModule.REPLACE_ONCE)
                self._ansible_filters[ansible_filter_key] = func

    def filters(self):
        """
        Returns the dictionary of available filters.
        :return: the dictionary of available filters.
        """
        return self._ansible_filters

    def sprig_abbrev(self, *args):
        raise NotImplementedError

    def sprig_abbrevboth(self, *args):
        raise NotImplementedError

    def sprig_add(self, *args):
        raise NotImplementedError

    def sprig_add1(self, *args):
        raise NotImplementedError

    def sprig_adler32sum(self, *args):
        raise NotImplementedError

    def sprig_ago(self, *args):
        raise NotImplementedError

    def sprig_append(self, *args):
        raise NotImplementedError

    def sprig_atoi(self, *args):
        raise NotImplementedError

    def sprig_b32dec(self, *args):
        raise NotImplementedError

    def sprig_b32enc(self, *args):
        raise NotImplementedError

    def sprig_b64dec(self, *args):
        raise NotImplementedError

    def sprig_b64enc(self, *args):
        raise NotImplementedError

    def sprig_base(self, *args):
        raise NotImplementedError

    def sprig_biggest(self, *args):
        raise NotImplementedError

    def sprig_buildCustomCert(self, *args):
        raise NotImplementedError

    def sprig_camelcase(self, *args):
        raise NotImplementedError

    def sprig_cat(self, *args):
        raise NotImplementedError

    def sprig_ceil(self, *args):
        raise NotImplementedError

    def sprig_clean(self, *args):
        raise NotImplementedError

    def sprig_coalesce(self, *args):
        raise NotImplementedError

    def sprig_compact(self, *args):
        raise NotImplementedError

    def sprig_concat(self, *args):
        raise NotImplementedError

    def sprig_contains(self, *args):
        raise NotImplementedError

    def sprig_date(self, *args):
        raise NotImplementedError

    def sprig_dateInZone(self, *args):
        raise NotImplementedError

    def sprig_dateModify(self, *args):
        raise NotImplementedError

    def sprig_date_in_zone(self, *args):
        raise NotImplementedError

    def sprig_date_modify(self, *args):
        raise NotImplementedError

    def sprig_decryptAES(self, *args):
        raise NotImplementedError

    def sprig_deepCopy(self, *args):
        raise NotImplementedError

    def sprig_deepEqual(self, *args):
        raise NotImplementedError

    def sprig_default(self, *args):
        raise NotImplementedError

    def sprig_derivePassword(self, *args):
        raise NotImplementedError

    def sprig_dict(self, *args):
        raise NotImplementedError

    def sprig_dir(self, *args):
        raise NotImplementedError

    def sprig_div(self, *args):
        raise NotImplementedError

    def sprig_empty(self, *args):
        raise NotImplementedError

    def sprig_encryptAES(self, *args):
        raise NotImplementedError

    def sprig_env(self, *args):
        raise NotImplementedError

    def sprig_expandenv(self, *args):
        raise NotImplementedError

    def sprig_ext(self, *args):
        raise NotImplementedError

    def sprig_fail(self, *args):
        raise NotImplementedError

    def sprig_first(self, *args):
        raise NotImplementedError

    def sprig_float64(self, *args):
        raise NotImplementedError

    def sprig_floor(self, *args):
        raise NotImplementedError

    def sprig_genCA(self, *args):
        raise NotImplementedError

    def sprig_genPrivateKey(self, *args):
        raise NotImplementedError

    def sprig_genSelfSignedCert(self, *args):
        raise NotImplementedError

    def sprig_genSignedCert(self, *args):
        raise NotImplementedError

    def sprig_getHostByName(self, *args):
        raise NotImplementedError

    def sprig_has(self, *args):
        raise NotImplementedError

    def sprig_hasKey(self, *args):
        raise NotImplementedError

    def sprig_hasPrefix(self, *args):
        raise NotImplementedError

    def sprig_hasSuffix(self, *args):
        raise NotImplementedError

    def sprig_hello(self, *args):
        raise NotImplementedError

    def sprig_htmlDate(self, *args):
        raise NotImplementedError

    def sprig_htmlDateInZone(self, *args):
        raise NotImplementedError

    def sprig_indent(self, *args):
        raise NotImplementedError

    def sprig_initial(self, *args):
        raise NotImplementedError

    def sprig_initials(self, *args):
        raise NotImplementedError

    def sprig_int(self, *args):
        raise NotImplementedError

    def sprig_int64(self, *args):
        raise NotImplementedError

    def sprig_isAbs(self, *args):
        raise NotImplementedError

    def sprig_join(self, *args):
        raise NotImplementedError

    def sprig_kebabcase(self, *args):
        raise NotImplementedError

    def sprig_keys(self, *args):
        raise NotImplementedError

    def sprig_kindIs(self, *args):
        raise NotImplementedError

    def sprig_kindOf(self, *args):
        raise NotImplementedError

    def sprig_last(self, *args):
        raise NotImplementedError

    def sprig_list(self, *args):
        raise NotImplementedError

    def sprig_lower(self, *args):
        raise NotImplementedError

    def sprig_max(self, *args):
        raise NotImplementedError

    def sprig_merge(self, *args):
        raise NotImplementedError

    def sprig_mergeOverwrite(self, *args):
        raise NotImplementedError

    def sprig_min(self, *args):
        raise NotImplementedError

    def sprig_mod(self, *args):
        raise NotImplementedError

    def sprig_mul(self, *args):
        raise NotImplementedError

    def sprig_nindent(self, *args):
        raise NotImplementedError

    def sprig_nospace(self, *args):
        raise NotImplementedError

    def sprig_now(self, *args):
        raise NotImplementedError

    def sprig_omit(self, *args):
        raise NotImplementedError

    def sprig_pick(self, *args):
        raise NotImplementedError

    def sprig_pluck(self, *args):
        raise NotImplementedError

    def sprig_plural(self, *args):
        raise NotImplementedError

    def sprig_prepend(self, *args):
        raise NotImplementedError

    def sprig_push(self, *args):
        raise NotImplementedError

    def sprig_quote(self, *args):
        raise NotImplementedError

    def sprig_randAlpha(self, *args):
        raise NotImplementedError

    def sprig_randAlphaNum(self, *args):
        raise NotImplementedError

    def sprig_randAscii(self, *args):
        raise NotImplementedError

    def sprig_randNumeric(self, *args):
        raise NotImplementedError

    def sprig_regexFind(self, *args):
        raise NotImplementedError

    def sprig_regexFindAll(self, *args):
        raise NotImplementedError

    def sprig_regexMatch(self, *args):
        raise NotImplementedError

    def sprig_regexReplaceAll(self, *args):
        raise NotImplementedError

    def sprig_regexReplaceAllLiteral(self, *args):
        raise NotImplementedError

    def sprig_regexSplit(self, *args):
        raise NotImplementedError

    def sprig_repeat(self, *args):
        raise NotImplementedError

    def sprig_replace(self, *args):
        raise NotImplementedError

    def sprig_rest(self, *args):
        raise NotImplementedError

    def sprig_reverse(self, *args):
        raise NotImplementedError

    def sprig_round(self, *args):
        raise NotImplementedError

    def sprig_semver(self, *args):
        raise NotImplementedError

    def sprig_semverCompare(self, *args):
        raise NotImplementedError

    def sprig_set(self, *args):
        raise NotImplementedError

    def sprig_sha1sum(self, *args):
        raise NotImplementedError

    def sprig_sha256sum(self, *args):
        raise NotImplementedError

    def sprig_shuffle(self, *args):
        raise NotImplementedError

    def sprig_slice(self, *args):
        raise NotImplementedError

    def sprig_snakecase(self, *args):
        raise NotImplementedError

    def sprig_sortAlpha(self, *args):
        raise NotImplementedError

    def sprig_split(self, *args):
        raise NotImplementedError

    def sprig_splitList(self, *args):
        raise NotImplementedError

    def sprig_splitn(self, *args):
        raise NotImplementedError

    def sprig_squote(self, *args):
        raise NotImplementedError

    def sprig_sub(self, *args):
        raise NotImplementedError

    def sprig_substr(self, *args):
        raise NotImplementedError

    def sprig_swapcase(self, *args):
        raise NotImplementedError

    def sprig_ternary(self, *args):
        raise NotImplementedError

    def sprig_title(self, *args):
        raise NotImplementedError

    def sprig_toDate(self, *args):
        raise NotImplementedError

    def sprig_toDecimal(self, *args):
        raise NotImplementedError

    def sprig_toJson(self, *args):
        raise NotImplementedError

    def sprig_toPrettyJson(self, *args):
        raise NotImplementedError

    def sprig_toString(self, *args):
        raise NotImplementedError

    def sprig_toStrings(self, *args):
        raise NotImplementedError

    def sprig_trim(self, *args):
        raise NotImplementedError

    def sprig_trimAll(self, *args):
        raise NotImplementedError

    def sprig_trimPrefix(self, *args):
        raise NotImplementedError

    def sprig_trimSuffix(self, *args):
        raise NotImplementedError

    def sprig_trimall(self, *args):
        raise NotImplementedError

    def sprig_trunc(self, *args):
        raise NotImplementedError

    def sprig_tuple(self, *args):
        raise NotImplementedError

    def sprig_typeIs(self, *args):
        raise NotImplementedError

    def sprig_typeIsLike(self, *args):
        raise NotImplementedError

    def sprig_typeOf(self, *args):
        raise NotImplementedError

    def sprig_uniq(self, *args):
        raise NotImplementedError

    def sprig_unixEpoch(self, *args):
        raise NotImplementedError

    def sprig_unset(self, *args):
        raise NotImplementedError

    def sprig_until(self, *args):
        raise NotImplementedError

    def sprig_untilStep(self, *args):
        raise NotImplementedError

    def sprig_untitle(self, *args):
        raise NotImplementedError

    def sprig_upper(self, *args):
        raise NotImplementedError

    def sprig_urlJoin(self, *args):
        raise NotImplementedError

    def sprig_urlParse(self, *args):
        raise NotImplementedError

    def sprig_uuidv4(self, *args):
        raise NotImplementedError

    def sprig_values(self, *args):
        raise NotImplementedError

    def sprig_without(self, *args):
        raise NotImplementedError

    def sprig_wrap(self, *args):
        raise NotImplementedError

    def sprig_wrapWith(self, *args):
        raise NotImplementedError
