import { CommandHandler } from "./common.ts";
import { wikiCommandHandler } from "./commands/wiki.ts";
import { pingjsCommandHandler } from "./commands/pingjs.ts";
import { sleepCommandHandler } from "./commands/sleep.ts";
import {
    APIChatInputApplicationCommandInteraction,
    APIInteraction,
    InteractionType,
} from "discord-api-types/v10";

export const commandHandlers: Record<string, CommandHandler> = {
    pingjs: pingjsCommandHandler,
    sleep: sleepCommandHandler,
    wiki: wikiCommandHandler,
};
export const isApplicationCommand = (
    interaction: APIInteraction,
): interaction is APIChatInputApplicationCommandInteraction =>
    interaction?.type === InteractionType.ApplicationCommand &&
    interaction?.data?.name in commandHandlers;
