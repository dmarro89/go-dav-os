"""Host-only providers for the Agent bridge service."""

import json
import urllib.error
import urllib.parse
import urllib.request

import fake_llm_bridge
from agent_bridge_planner import parse_provider_content, provider_messages


MAX_PROVIDER_BODY = 65536


class ProviderError(Exception):
    def __init__(self, message, code="provider_error"):
        super().__init__(message)
        self.code = code


class Provider:
    def plan(self, request):
        raise NotImplementedError


class FakeProvider(Provider):
    def plan(self, request):
        return fake_llm_bridge.plan_for_request(request)


class OpenAICompatibleProvider(Provider):
    def __init__(self, api_key, model, base_url, timeout=10.0):
        try:
            scheme = urllib.parse.urlsplit(base_url).scheme.lower()
        except ValueError as error:
            raise ProviderError(
                "provider endpoint must use http or https", "provider_config"
            ) from error
        if scheme not in ("http", "https"):
            raise ProviderError(
                "provider endpoint must use http or https", "provider_config"
            )
        self.api_key = api_key
        self.model = model
        self.endpoint = base_url.rstrip("/") + "/chat/completions"
        self.timeout = timeout

    def plan(self, request):
        body = json.dumps(
            {
                "model": self.model,
                "messages": provider_messages(request),
            }
        ).encode("utf-8")
        try:
            provider_request = urllib.request.Request(
                self.endpoint,
                data=body,
                headers={
                    "Authorization": "Bearer " + self.api_key,
                    "Content-Type": "application/json",
                },
                method="POST",
            )
            with urllib.request.urlopen(provider_request, timeout=self.timeout) as response:
                raw_response = response.read(MAX_PROVIDER_BODY + 1)
        except (OSError, TimeoutError, ValueError, urllib.error.URLError) as error:
            raise ProviderError("provider request failed") from error
        if len(raw_response) > MAX_PROVIDER_BODY:
            raise ProviderError("provider response is too large")

        try:
            envelope = json.loads(raw_response.decode("utf-8"))
            content = envelope["choices"][0]["message"]["content"]
        except (
            KeyError,
            IndexError,
            TypeError,
            UnicodeDecodeError,
            json.JSONDecodeError,
            RecursionError,
        ) as error:
            raise ProviderError("provider returned an invalid response") from error
        return parse_provider_content(content)


def provider_from_environment(environment):
    provider_name = (environment.get("DAVOS_AGENT_PROVIDER", "") or "").strip().lower()
    provider_name = provider_name or "fake"
    if provider_name == "fake":
        return FakeProvider()
    if provider_name != "openai":
        raise ProviderError("unsupported provider", "provider_config")

    api_key = environment.get("DAVOS_AGENT_API_KEY", "")
    model = environment.get("DAVOS_AGENT_MODEL", "")
    base_url = environment.get("DAVOS_AGENT_BASE_URL", "")
    if not api_key or not model or not base_url:
        raise ProviderError("openai provider configuration is incomplete", "provider_config")
    return OpenAICompatibleProvider(api_key, model, base_url)
