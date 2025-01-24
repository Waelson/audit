import React, { useState, useEffect } from "react";
import "./App.css";
import axios from "axios";

const App = () => {
  // Função para gerar valores aleatórios
  const generateRandomFormData = () => {
    const randomOrderNumber = `ORD${Math.floor(Math.random() * 1000000)}`;
    const randomPaymentAmount = (Math.random() * 1000).toFixed(2);
    const randomTransactionAmount = randomPaymentAmount; // Pode ser ajustado se necessário
    const randomName = `Name ${Math.floor(Math.random() * 1000)}`;
    const randomCardNumber = Array(16)
        .fill(0)
        .map(() => Math.floor(Math.random() * 10))
        .join("");
    const randomExpiryDate = `${String(Math.floor(Math.random() * 12) + 1).padStart(2, "0")}/${
        Math.floor(Math.random() * 10) + 25
    }`;
    const randomSecurityCode = String(Math.floor(Math.random() * 1000)).padStart(3, "0");
    const randomPostalCode = String(Math.floor(Math.random() * 10000)).padStart(5, "0");
    const randomTransactionDateTime = new Date().toISOString();

    return {
      orderNumber: randomOrderNumber,
      paymentAmount: parseFloat(randomPaymentAmount),
      transactionAmount: parseFloat(randomTransactionAmount),
      nameOnCard: randomName,
      cardNumber: randomCardNumber,
      expiryDate: randomExpiryDate,
      securityCode: randomSecurityCode,
      postalCode: randomPostalCode,
      transactionDateTime: randomTransactionDateTime,
    };
  };

  // Estado inicial com dados aleatórios
  const [formData, setFormData] = useState(generateRandomFormData);

  // Gera novos dados aleatórios após o envio da requisição
  const resetFormData = () => {
    const newFormData = generateRandomFormData();
    console.log("[INFO] Novos dados gerados para o formulário:", newFormData);
    setFormData(newFormData);
  };

  // Gera novos dados aleatórios ao carregar o formulário
  useEffect(() => {
    console.log("[INFO] Formulário carregado com dados aleatórios:", formData);
  }, []);

  const handleChange = (e) => {
    const { name, value } = e.target;
    console.log(`[INFO] Atualizando campo ${name} com valor: ${value}`);
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    console.log("[INFO] Tentando enviar os dados de pagamento...");
    console.log("[DEBUG] Dados do formulário enviados:", formData);

    try {
      const response = await axios.post("http://localhost:8080/api/v1/payment", formData);
      console.log("[INFO] Pagamento processado com sucesso:", response.data);
      alert("Payment successful!");
      resetFormData(); // Gera novos dados aleatórios
    } catch (error) {
      console.error("[ERROR] Falha ao processar o pagamento:", error);
      alert("Payment failed. Please try again.");
    }
  };

  return (
      <div className="payment-container">
        <h1>Pay Invoice</h1>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Order Number</label>
            <input
                type="text"
                name="orderNumber"
                value={formData.orderNumber}
                onChange={handleChange}
                required
            />
          </div>
          <div className="form-group">
            <label>Payment amount</label>
            <div>
              <span>${formData.paymentAmount.toFixed(2)}</span>
            </div>
          </div>
          <div className="form-group">
            <label>Transaction amount</label>
            <input
                type="number"
                name="transactionAmount"
                value={formData.transactionAmount}
                onChange={handleChange}
                required
            />
          </div>
          <div className="form-group">
            <label>Name on card</label>
            <input
                type="text"
                name="nameOnCard"
                value={formData.nameOnCard}
                onChange={handleChange}
                required
            />
          </div>
          <div className="form-group">
            <label>Card number</label>
            <input
                type="text"
                name="cardNumber"
                value={formData.cardNumber}
                onChange={handleChange}
                required
            />
          </div>
          <div className="form-group">
            <label>Expiry date</label>
            <input
                type="text"
                name="expiryDate"
                placeholder="MM / YY"
                value={formData.expiryDate}
                onChange={handleChange}
                required
            />
          </div>
          <div className="form-group">
            <label>Security code</label>
            <input
                type="text"
                name="securityCode"
                value={formData.securityCode}
                onChange={handleChange}
                required
            />
          </div>
          <div className="form-group">
            <label>ZIP/Postal code</label>
            <input
                type="text"
                name="postalCode"
                value={formData.postalCode}
                onChange={handleChange}
                required
            />
          </div>
          <div className="form-group">
            <label>Transaction DateTime</label>
            <input
                type="datetime-local"
                name="transactionDateTime"
                value={formData.transactionDateTime}
                onChange={(e) => {
                  console.log("[INFO] Atualizando transactionDateTime com valor:", e.target.value);
                  setFormData({ ...formData, transactionDateTime: e.target.value });
                }}
                required
            />
          </div>
          <button type="submit" className="pay-button">
            Pay ${formData.paymentAmount.toFixed(2)}
          </button>
        </form>
      </div>
  );
};

export default App;
