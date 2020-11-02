using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using Braintree;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Infrastructure;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

// For more information on enabling Web API for empty projects, visit https://go.microsoft.com/fwlink/?LinkID=397860

namespace CharityYetiServices.Controllers
{
    [Route("api/[controller]")]
    [ApiController]
    public class PaymentController : ControllerBase
    {
        private readonly ILogger<PaymentController> _logger;
        private readonly BraintreeGateway _gateway;
        private readonly IConfiguration _configuration;

        public PaymentController(ILogger<PaymentController> logger, IConfiguration configuration)
        {
            _logger = logger;
            _configuration = configuration;

            // TODO: Tokenize in config file
            _gateway = new BraintreeGateway
            {
                Environment = Braintree.Environment.SANDBOX,
                MerchantId = _configuration["Braintree:MerchantID"],
                PublicKey = _configuration["Braintree:PublicKey"],
                PrivateKey = _configuration["Braintree:PrivateKey"],
            };
        }

        [HttpGet("healthcheck")]
        public ActionResult HealthCheck()
        {
            return Ok("Healthy AF");
        }

        // GET api/<PaymentController>/GetClientToken
        [HttpGet]
        public string InitiatePayment()
        {
            return InitiatePayment(null);
        }

        // GET api/<PaymentController>/GetClientToken/5
        [HttpGet("{id}")]
        public string InitiatePayment(string customerId)
        {
            // Generate a BT client token
            string clientToken = String.Empty;

            try
            {
                clientToken = GetClientTokenFromBraintree(customerId);
            }

            catch (Exception ex)
            {
                _logger.LogError(ex, ex.Message);
                return Convert.ToString(new StatusCodeResult(StatusCodes.Status500InternalServerError));
            }

            // Pass clientToken to the front-end so it can generate a Nonce
            return clientToken;
        }

        [HttpPost]
        public ActionResult CreatePurchase(Dictionary<string, string> orderData)
        {
            string nonceFromTheClient = orderData["payment_method_nonce"] ?? orderData["payment_method_nonce"];
            string deviceDataFromTheClient = orderData["device_data"] ?? orderData["device_data"];
            decimal paymentAmount = Convert.ToDecimal(orderData["payment_amount"] ?? orderData["payment_amount"]);

            // Use payment method nonce here
            var request = new TransactionRequest
            {
                Amount = paymentAmount,
                PaymentMethodNonce = nonceFromTheClient,
                DeviceData = deviceDataFromTheClient,
                Options = new TransactionOptionsRequest
                {
                    SubmitForSettlement = true
                }
            };

            Result<Transaction> result = _gateway.Transaction.Sale(request);

            if (result.IsSuccess())
            {
                // Woo!
                return Ok(result.Transaction); // Return the entire transaction data for the business layer
            }

            else
            {
                // Do something with errors
                foreach (var error in result.Errors.All())
                    _logger.LogError($"Transaction error: {error.Attribute} {error.Code} {error.Message}");

                throw new Exception($"There were {result.Errors.Count} error/s with the transaction");
            }
        }

        /// <summary>
        /// Gets the client token from Braintree
        /// </summary>
        /// <param name="customerId">Optional to get a customer-specific client token for some improved UX features</param>
        /// <returns></returns>
        private string GetClientTokenFromBraintree(string customerId)
        {
            if (!String.IsNullOrWhiteSpace(customerId))
            {
                try
                {
                    var clientToken = _gateway.ClientToken.Generate(
                        new ClientTokenRequest
                        {
                            CustomerId = customerId
                        }
                    );

                    return clientToken;
                }

                catch (ArgumentException ex)
                {
                    _logger.LogError(ex, $"Client ID: {customerId} Error: {ex.Message}");
                }
            }

            return _gateway.ClientToken.Generate();
        }
    }
}
