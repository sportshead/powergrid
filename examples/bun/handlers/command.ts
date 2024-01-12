import { CommandHandler } from "./common.ts";
import { wikiCommandHandler } from "./commands/wiki.ts";
import { pingjsCommandHandler } from "./commands/pingjs.ts";
import { sleepCommandHandler } from "./commands/sleep.ts";
import {
    APIChatInputApplicationCommandInteraction,
    APIInteraction,
    InteractionType,
} from "discord-api-types/v10";
import { counterCommandHandler } from "./commands/counter.ts";

export const commandHandlers: Record<string, CommandHandler> = {
    pingjs: pingjsCommandHandler,
    sleep: sleepCommandHandler,
    wiki: wikiCommandHandler,
    counter: counterCommandHandler,
};
export const isApplicationCommand = (
    interaction: APIInteraction,
): interaction is APIChatInputApplicationCommandInteraction =>
    interaction?.type === InteractionType.ApplicationCommand &&
    interaction?.data?.name in commandHandlers;
