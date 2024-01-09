import { AutocompleteHandler } from "./common.ts";
import { wikiAutocompleteHandler } from "./commands/wiki.ts";
import {
    APIApplicationCommandAutocompleteInteraction,
    APIInteraction,
    InteractionType,
} from "discord-api-types/v10";

export const autocompleteHandlers: Record<string, AutocompleteHandler> = {
    wiki: wikiAutocompleteHandler,
};

export const isAutocomplete = (
    interaction: APIInteraction,
): interaction is APIApplicationCommandAutocompleteInteraction =>
    interaction?.type === InteractionType.ApplicationCommandAutocomplete &&
    interaction?.data?.name in autocompleteHandlers;
