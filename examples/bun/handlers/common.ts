import {
    APIApplicationCommandAutocompleteInteraction,
    APIChatInputApplicationCommandInteraction,
} from "discord-api-types/v10";
import type { APIChatInputApplicationCommandInteractionData } from "discord-api-types/payloads/v10";

export const getOption = <T>(
    interaction: {
        data: Pick<APIChatInputApplicationCommandInteractionData, "options">;
    },
    name: string,
    defaultValue: T,
) =>
    (
        interaction.data.options?.find((o) => o.name === name) as
            | { value: T | undefined }
            | undefined
    )?.value ?? defaultValue;

// https://discord.com/developers/docs/reference#snowflakes-snowflake-id-format-structure-left-to-right
const DISCORD_EPOCH = 1420070400000;
export const snowflakeToDate = (snowflake: number) =>
    new Date(Number(BigInt(snowflake) >> 22n) + DISCORD_EPOCH);
export const dateToTimestamp = (d: Date) => Math.floor(d.getTime() / 1000);

export type CommandHandler = (
    interaction: APIChatInputApplicationCommandInteraction,
) => Promise<Response>;
export type AutocompleteHandler = (
    interaction: APIApplicationCommandAutocompleteInteraction,
) => Promise<Response>;
