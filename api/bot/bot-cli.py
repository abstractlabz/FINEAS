import discord
from discord.ext import commands
import requests
import os
import urllib.parse
from dotenv import load_dotenv
from discord.ext.commands import DefaultHelpCommand
import re
import hashlib

load_dotenv()

DISCORD_TOKEN = os.getenv('DISCORD_TOKEN') or ""
RETRIEVAL_API_URL = os.getenv('RETRIEVAL_API_URL')
CHATBOT_API_URL = os.getenv('CHATBOT_API_URL')
AUTH_BEARER = os.getenv('AUTH_BEARER')
UPGRADE_API_URL = os.getenv('UPGRADE_API_URL')

bot = commands.Bot(command_prefix='!', intents=discord.Intents.all())

class CustomHelpCommand(DefaultHelpCommand):
    async def send_bot_help(self, mapping):
        embed = discord.Embed(title="Help - Available Commands", color=discord.Color.blue())
        for cog, commands in mapping.items():
            for command in commands:
                embed.add_field(name=f"!{command.name}", value=command.help, inline=False)
        await self.get_destination().send(embed=embed)

bot.help_command = CustomHelpCommand()

@bot.event
async def on_ready():
    print(f'{bot.user} has connected to Discord!')

def hash_user_id(user_id):
    return hashlib.sha256(user_id.encode()).hexdigest()

def chunk_text(text, max_chars=1024):
    paragraphs = text.split('\n')
    chunks = []
    current_chunk = ""
    current_char_count = 0

    for paragraph in paragraphs:
        if current_char_count + len(paragraph) + 1 > max_chars:
            chunks.append(current_chunk)
            current_chunk = paragraph
            current_char_count = len(paragraph) + 1
        else:
            if current_chunk:
                current_chunk += "\n" + paragraph
            else:
                current_chunk = paragraph
            current_char_count += len(paragraph) + 1

    if current_chunk:
        chunks.append(current_chunk)

    return chunks

async def send_embed(ctx, title, content):
    chunks = chunk_text(content)
    for chunk in chunks:
        embed = discord.Embed(title=title, description=chunk, color=discord.Color.blue())
        await ctx.send(embed=embed)

async def get_user_info(ctx):
    user_id = str(ctx.author.id)
    hashed_id = hash_user_id(user_id)
    print(hashed_id)
    response = requests.get(f"{UPGRADE_API_URL}/get-user-info", params={"id_hash": hashed_id, "email": None})
    if response.status_code != 200:
        await ctx.send("Error retrieving user info. Please try again later.")
        return None
    return response.json()

async def enforce_credits(ctx):
    user_info = await get_user_info(ctx)
    if not user_info:
        return False, None
    user_id = str(ctx.author.id)
    hashed_id = hash_user_id(user_id)
    response = requests.post(f"{UPGRADE_API_URL}/enforce-credits", json={"id_hash": hashed_id})
    if response.status_code == 402:
        await ctx.send("You have run out of credits. Please upgrade your membership.")
        return False, None
    elif response.status_code != 200:
        await ctx.send("Error enforcing credits. Please try again later.")
        return False, None
    remaining_credits = response.json().get('credits', None)
    return True, remaining_credits

async def notify_credits(ctx, remaining_credits):
    if remaining_credits is not None:
        await send_embed(ctx, "", f"You have {remaining_credits} credits remaining.")


@bot.command(name='stk', help='Get stock information about a specific ticker')
async def stk(ctx: commands.Context, symbol: str) -> None:
    success, remaining_credits = await enforce_credits(ctx)
    if not success:
        return
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        data = response.json()
        stock_performance = data.get('StockPerformance', 'N/A')
        await send_embed(ctx, f"Information for {symbol}", stock_performance)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")
    await notify_credits(ctx, remaining_credits)

@bot.command(name='fin', help='Get financial information about a specific ticker')
async def fin(ctx: commands.Context, symbol: str) -> None:
    success, remaining_credits = await enforce_credits(ctx)
    if not success:
        return
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        data = response.json()
        financial_health = data.get('FinancialHealth', 'N/A')
        await send_embed(ctx, f"Information for {symbol}", financial_health)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")
    await notify_credits(ctx, remaining_credits)

@bot.command(name='news', help='Get news summaries about a specific ticker')
async def news(ctx: commands.Context, symbol: str) -> None:
    success, remaining_credits = await enforce_credits(ctx)
    if not success:
        return
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        data = response.json()
        news_summary = data.get('NewsSummary', 'N/A')
        await send_embed(ctx, f"Information for {symbol}", news_summary)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")
    await notify_credits(ctx, remaining_credits)

@bot.command(name='desc', help='Get descriptions about a specific ticker')
async def desc(ctx: commands.Context, symbol: str) -> None:
    success, remaining_credits = await enforce_credits(ctx)
    if not success:
        return
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        data = response.json()
        company_desc = data.get('CompanyDesc', 'N/A')
        await send_embed(ctx, f"Information for {symbol}", company_desc)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")
    await notify_credits(ctx, remaining_credits)

@bot.command(name='ta', help='Get technical analysis for a specific ticker')
async def ta(ctx: commands.Context, symbol: str) -> None:
    success, remaining_credits = await enforce_credits(ctx)
    if not success:
        return
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        data = response.json()
        technical_analysis = data.get('TechnicalAnalysis', 'N/A')
        await send_embed(ctx, f"Information for {symbol}", technical_analysis)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")
    await notify_credits(ctx, remaining_credits)

@bot.command(name='ask', help='Ask an investment research question')
async def ask(ctx: commands.Context, *, question: str) -> None:
    success, remaining_credits = await enforce_credits(ctx)
    if not success:
        return
    try:
        question = urllib.parse.quote(question)
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.post(f"{CHATBOT_API_URL}/chat?prompt={question}", headers=headers)
        answer = response.text
        await send_embed(ctx, "Answer to your question", answer)
    except Exception as e:
        await ctx.send("Error processing your question. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")
    await notify_credits(ctx, remaining_credits)

@bot.command(name='checkout', help='Generate a Stripe checkout link')
async def checkout(ctx: commands.Context) -> None:
    user_id = str(ctx.author.id)
    hashed_id = hash_user_id(user_id)
    response = requests.post(f"{UPGRADE_API_URL}/upgrade_membership", json={"id_hash": hashed_id, "email": None})
    if response.status_code == 200:
        data = response.json()
        checkoutSessionURL = data['url']
        await ctx.send(f"Checkout link: {checkoutSessionURL}")
    else:
        await ctx.send("Error generating checkout link. Please try again later.")

@bot.command(name='credits', help='Check your remaining credits')
async def credits(ctx: commands.Context) -> None:
    user_id = str(ctx.author.id)
    hashed_id = hash_user_id(user_id)
    response = requests.get(f"{UPGRADE_API_URL}/get-user-info", params={"id_hash": hashed_id})
    if response.status_code == 200:
        data = response.json()
        await send_embed(ctx, "", f"You have {data['user']['credits']} credits remaining")
    else:
        await send_embed(ctx, "", f"Error fetching credits")


@bot.command(name='cancel', help='Cancel your Stripe subscription')
async def cancel(ctx: commands.Context) -> None:
    user_id = str(ctx.author.id)
    hashed_id = hash_user_id(user_id)
    response = requests.post(f"{UPGRADE_API_URL}/cancel-subscription", json={"id_hash": hashed_id})
    if response.status_code == 200:
        await ctx.send("Your subscription has been canceled.")
    else:
        await ctx.send("Error canceling subscription. Please try again later.")

if DISCORD_TOKEN:
    bot.run(DISCORD_TOKEN)
else:
    print("Error: DISCORD_TOKEN is not set.")