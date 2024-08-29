import discord
from discord.ext import commands
import requests
import os
import urllib.parse
from dotenv import load_dotenv
from discord.ext.commands import DefaultHelpCommand

load_dotenv()

DISCORD_TOKEN = os.getenv('DISCORD_TOKEN') or ""
RETRIEVAL_API_URL = os.getenv('RETRIEVAL_API_URL')
CHATBOT_API_URL = os.getenv('CHATBOT_API_URL')
AUTH_BEARER = os.getenv('AUTH_BEARER')

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

def chunk_text(text, chunk_size=1024):
    return [text[i:i + chunk_size] for i in range(0, len(text), chunk_size)]

@bot.command(name='stk', help='Get stock information about a specific ticker')
async def stk(ctx: commands.Context, symbol: str) -> None:
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        print(f"Response Status Code: {response.status_code}")
        print(f"Response Content: {response.content}")
        data = response.json()
        embed = discord.Embed(title=f"Information for {symbol}", color=discord.Color.blue())
        stock_performance = data.get('StockPerformance', 'N/A')
        for chunk in chunk_text(stock_performance):
            embed.add_field(name="Stock Performance", value=chunk, inline=False)
        await ctx.send(embed=embed)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")

@bot.command(name='fin', help='Get financial information about a specific ticker')
async def fin(ctx: commands.Context, symbol: str) -> None:
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        print(f"Response Status Code: {response.status_code}")
        print(f"Response Content: {response.content}")
        data = response.json()
        embed = discord.Embed(title=f"Information for {symbol}", color=discord.Color.blue())
        financial_health = data.get('FinancialHealth', 'N/A')
        for chunk in chunk_text(financial_health):
            embed.add_field(name="Financial Health", value=chunk, inline=False)
        await ctx.send(embed=embed)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")

@bot.command(name='news', help='Get news summaries about a specific ticker')
async def news(ctx: commands.Context, symbol: str) -> None:
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        print(f"Response Status Code: {response.status_code}")
        print(f"Response Content: {response.content}")
        data = response.json()
        embed = discord.Embed(title=f"Information for {symbol}", color=discord.Color.blue())
        news_summary = data.get('NewsSummary', 'N/A')
        for chunk in chunk_text(news_summary):
            embed.add_field(name="News Summary", value=chunk, inline=False)
        await ctx.send(embed=embed)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")

@bot.command(name='desc', help='Get descriptions about a specific ticker')
async def desc(ctx: commands.Context, symbol: str) -> None:
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        print(f"Response Status Code: {response.status_code}")
        print(f"Response Content: {response.content}")
        data = response.json()
        embed = discord.Embed(title=f"Information for {symbol}", color=discord.Color.blue())
        company_desc = data.get('CompanyDesc', 'N/A')
        for chunk in chunk_text(company_desc):
            embed.add_field(name="Company Description", value=chunk, inline=False)
        await ctx.send(embed=embed)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")

@bot.command(name='ta', help='Get technical analysis for a specific ticker')
async def ta(ctx: commands.Context, symbol: str) -> None:
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        print(f"Response Status Code: {response.status_code}")
        print(f"Response Content: {response.content}")
        data = response.json()
        embed = discord.Embed(title=f"Information for {symbol}", color=discord.Color.blue())
        technical_analysis = data.get('TechnicalAnalysis', 'N/A')
        for chunk in chunk_text(technical_analysis):
            embed.add_field(name="Technical Analysis", value=chunk, inline=False)
        await ctx.send(embed=embed)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")

@bot.command(name='ask', help='Ask an investment research question')
async def ask(ctx: commands.Context, *, question: str) -> None:
    try:
        question = urllib.parse.quote(question)
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.post(f"{CHATBOT_API_URL}/chat?prompt={question}", headers=headers)
        answer = response.text
        embed = discord.Embed(title="Fineas Says...", color=discord.Color.blue())
        for chunk in chunk_text(answer):
            embed.add_field(name="Response", value=chunk, inline=False)
        await ctx.send(embed=embed)
    except Exception as e:
        await ctx.send("Error processing your question. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")

if DISCORD_TOKEN:
    bot.run(DISCORD_TOKEN)
else:
    print("Error: DISCORD_TOKEN is not set.")